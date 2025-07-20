package handlers

import (
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	tflite "github.com/mattn/go-tflite"
	"github.com/nfnt/resize"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	// "golang.org/x/text/number"
)

func GetFour(c *fiber.Ctx) error {
    // Tạo context từ Fiber
	ctx := c.Context()

	// Truy vấn danh sách sản phẩm
	products, err := models.Products(qm.Limit(4)).All(ctx, boil.GetContextDB())
	if err != nil {
		return service.SendError(c,500,"Failed to fetch products")
	}

	resp := fiber.Map{
		"status": "Success",
		"data": products,
		"message": "Successfully fetched featuring items",
	}

	// Trả về danh sách sản phẩm dưới dạng JSON
	return c.JSON(resp)
}

func GetTypes(c *fiber.Ctx) error{
	types,err := models.ProductTypes().All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,"Failed to fetch product types!")
	}
	resp := fiber.Map{
		"status":"Success",
		"data": types,
		"message": "Successfully fetched product types",
	}
	return c.JSON(resp);
}

func GetProductsByPage(c *fiber.Ctx) error{
	var totalProduct int64

	//Lấy về số trang
	page,err := strconv.Atoi(c.Query("page","1"));
	if err != nil{
		return service.SendError(c,500,"Error converting pageNum");
	}

	//Lấy về các tham số
	typeName := c.Query("type","");
	search := c.Query("search","");
	minPrice := c.QueryInt("minPrice",0);
	maxPrice := c.QueryInt("maxPrice",0);
	orderBy := c.Query("orderBy","ASC");
	//Tính offset
	offset := (page-1)*6;

	//Tạo query mod
	queryMods := []qm.QueryMod{
		qm.Where("price BETWEEN ? AND ?",minPrice,maxPrice),
		qm.Load(models.ProductRels.ProductTypeIDProductType),
	}

	//Nếu có typeName thì lấy totalProduct theo typeName
	if typeName != ""{
		productType,err := models.ProductTypes(qm.Where("\"typeName\" = ?",typeName)).One(c.Context(),boil.GetContextDB())
		if err != nil {
			return service.SendError(c, 500, "Product type not found")
		}
		queryMods = append(queryMods, qm.Where( "\"productTypeID\" = ?",productType.ProductTypeID),)
		totalProduct,err = models.Products(
			queryMods...
		).Count(c.Context(),boil.GetContextDB());
		if err != nil {
			return service.SendError(c, 500, "Total product not found")
		}
	}
	//Nếu có search thì cũng làm tương tự
	if search != ""{
		queryMods = append(queryMods, qm.Where("LOWER(\"productName\") LIKE LOWER(?)","%"+search+"%"))
		totalProduct,err = models.Products(
			queryMods...
		).Count(c.Context(),boil.GetContextDB())
		if err != nil {
			return service.SendError(c, 500, "Total product not found")
		}
	}

	//Không có typeName hay search thì lấy hết tất cả totalProduct
	totalProduct,err = models.Products(
		queryMods...
	).Count(c.Context(),boil.GetContextDB());
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}
	totalPage := int(math.Ceil(float64(totalProduct) / float64(6)))

	//Creating redis key after page,type,search
	redisKey := fmt.Sprintf("products:page%d:type=%s:search=%s:minPrice=%d:maxPrice=%d:orderByPrice=%s",page,typeName,search,minPrice,maxPrice,orderBy)
	//Checking if redis key exists
	cachedProducts,err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result()
	fmt.Println("Cached data:", cachedProducts)
	if err == nil{
		return c.JSON(json.RawMessage(cachedProducts))
	}

	//Thêm vào offset và limit để phân trang
	queryMods = append(queryMods,qm.OrderBy("price "+orderBy), qm.Limit(6), qm.Offset(offset));
	products, err := models.Products(queryMods...).All(c.Context(), boil.GetContextDB())
	
	
	if err != nil {
		println(err.Error())
		return service.SendError(c, 500, "Failed to fetch products by page")
	}

	resp := fiber.Map{
		"status": "Success",
		"data": products,
		"totalPage": totalPage,
		"message": "Successfully fetched products by page",
	}

	//saving redis key to redis database for 10 mins
	jsonData, _ := json.Marshal(resp)
	rdsErr := redisdatabase.Client.Set(redisdatabase.Ctx,redisKey,jsonData, 10*time.Minute)
	if rdsErr != nil{
		fmt.Println("Failed to cache product data:", rdsErr)
	}


	return c.JSON(resp);
}

var classNames []string = []string{
	"Bánh flan", "Bánh mì ngọt", "Bánh mochi", "Tiramisu",
	"Chè thái", "Cơm bò lúc lắc", "Cơm cá chiên", "Cơm chiên dương châu", "Cơm gà", "Cơm tấm",
	"Cơm thịt kho", "Cơm xá xíu", "Kem dừa", "Kem socola", "Nước ngọt 7up", "Nước ngọt coca-cola",
	"Nước ngọt pepsi", "Nước ngọt sprite", "Nước tăng lực red bull", "Nước tăng lực sting", "Thịt bò hầm tiêu xanh",
	"Thịt heo quay",
}

func ClassifyImage(c *fiber.Ctx) error{
	//Load model
	modelPath := "internal/models/model_unquant.tflite";
	model := tflite.NewModelFromFile(modelPath);
	if model == nil{
		return service.SendError(c,500,"Failed to load TFLite model");
	}

	//Tạo interpreter
	defer model.Delete()
	options := tflite.NewInterpreterOptions()
	defer options.Delete()
	interpreter := tflite.NewInterpreter(model,options)
	defer interpreter.Delete()

	//Cấp phát bộ nhớ cho tensors
	interpreter.AllocateTensors()

	//Lấy dữ liệu ảnh từ request
	file, err := c.FormFile("image")
	if err != nil{
		return service.SendError(c,400,"Invalid request format");
	}

	// Mở file ảnh
	fileContent, err := file.Open()
	if err != nil {
		return service.SendError(c, 500, "Failed to open uploaded image")
	}
	defer fileContent.Close()

	// Đọc dữ liệu ảnh vào buffer
	imageBytes, err := io.ReadAll(fileContent)
	if err != nil {
		return service.SendError(c, 500, "Failed to read image data")
	}

	//Chuyển ảnh thành tensor
	tensorData,err := processImageForModel(imageBytes);
	if err != nil{
		fmt.Println("Error in processImageForModel:", err)
		return service.SendError(c,500,"Failed to process image for model");
	}

	//Gán dữ liệu tensor vào input
	inputTensor := interpreter.GetInputTensor(0)
	copy(inputTensor.Float32s(),tensorData)

	//Thực hiện inference
	if err := interpreter.Invoke(); err != tflite.Error{
		return service.SendError(c,500,"Failed to invoke model");
	}

	//Lấy kết quả từ output tensor
	outputTensor := interpreter.GetOutputTensor(0)
	logits := outputTensor.Float32s()

	//Áp dụng softmax để tính xác xuất
	probabilities := softmax(logits)

	//Mapping className với xác suất
	resultWithConfidence := make([]map[string]interface{},len(probabilities))
	for i,prob := range probabilities{
		if i < len(classNames){
			resultWithConfidence[i] = map[string]interface{}{
				"className": classNames[i],
				"confidence": prob,
			}
		}
	}

	//Sắp xếp kết quả theo độ tin cậy
	sort.Slice(resultWithConfidence, func(i, j int) bool{
		return resultWithConfidence[i]["confidence"].(float32) > resultWithConfidence[j]["confidence"].(float32)
	})

	// Giữ lại 3 kết quả có độ tin cậy cao nhất
	top3Results := resultWithConfidence[:3]

	//Trả về kết quả
	return c.JSON(fiber.Map{
		"status": "Success",
		"message": "Image classified successfully",
		"data": top3Results,
	})

}

func processImageForModel(imageBytes []byte) ([]float32, error) {
	// Kiểm tra kích thước dữ liệu ảnh
	fmt.Println("Image byte length:", len(imageBytes))

	// Thử decode ảnh bằng các thư viện cụ thể
	img, format, err := image.Decode(bytes.NewReader(imageBytes))
	fmt.Println(format)
	if err != nil {
		fmt.Println("Failed to decode image using image.Decode. Trying alternative decoders...")

		// Thử giải mã ảnh JPEG
		img, err = jpeg.Decode(bytes.NewReader(imageBytes))
		if err != nil {
			fmt.Println("JPEG decode failed, trying PNG...")
			// Thử giải mã ảnh PNG
			img, err = png.Decode(bytes.NewReader(imageBytes))
		}

		if err != nil {
			return nil, fmt.Errorf("failed to decode image using all formats: %v", err)
		}
	}

	// Resize ảnh về 224x224
	resizedImg := resize.Resize(224, 224, img, resize.Lanczos3)

	// Chuyển ảnh thành dữ liệu tensor
	tensorData := make([]float32, 224*224*3)
	index := 0
	for y := 0; y < 224; y++ {
		for x := 0; x < 224; x++ {
			r, g, b, _ := resizedImg.At(x, y).RGBA()
			tensorData[index] = float32(r>>8) / 255.0 // Chuẩn hóa pixel về [0,1]
			tensorData[index+1] = float32(g>>8) / 255.0
			tensorData[index+2] = float32(b>>8) / 255.0
			index += 3
		}
	}

	return tensorData, nil
}

func softmax(logits []float32) []float32{
	exp := make([]float32, len(logits))
	var sum float32
	for i, logit := range logits{
		exp[i] = float32(math.Exp(float64(logit)))
		sum += exp[i]
	}
	for i := range exp{
		exp[i] /= sum
	}
	return exp
}

type Star struct{
	FiveStars int `json:"fiveStars"`
	FourStars int `json:"fourStars"`
	ThreeStars int `json:"threeStars"`
	TwoStars int `json:"twoStars"`
	OneStars int `json:"oneStars"`
}
type ProductDetailResponse struct{
	models.Product `json:"product"`
	FiveStarsReview []ReviewResponse `json:"review"`
	Stars Star `json:"stars"`
}
func GetDetail(c *fiber.Ctx) error{
	detailedResponse := ProductDetailResponse{}
	id := c.QueryInt("id",0);
	if id == 0{
		return service.SendError(c,500,"ID not found");
	}
	filter := c.Query("filter","All");
	page := c.QueryInt("page",1);
	offset := (page - 1) * 3;

	//creating redis key
	redisKey := fmt.Sprintf("product:detail:%d:filter=%s:page=%d",id,filter,page)
	//checking if redis key exists
	cachedDetail,err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result()
	if err == nil{
		return c.JSON(json.RawMessage(cachedDetail))
	}
	
	//fetching and setting data of ProductDetailResponse
	detail, err := models.Products(qm.Where("\"productID\" = ?",id)).One(c.Context(),boil.GetContextDB());
	if err != nil {
		return service.SendError(c, 500, "product not found");
	}
	detailedResponse.Product = *detail

	//filter logic
	reviewResult, err, totalPage := reviewDisplay(c.Context(),id, filter,offset);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	detailedResponse.FiveStarsReview = reviewResult

	//nên gắn key params vào cho redis key vì lúc này cần phân biết giưa các redis key với nhau
	
	//counting total number of reviews for each rating	
	fiveStars, _ := models.Reviews(qm.Where("\"productID\" = ? AND stars = 5",id)).Count(c.Context(),boil.GetContextDB())
	detailedResponse.Stars.FiveStars = int(fiveStars)
	fourStars, _ := models.Reviews(qm.Where("\"productID\" = ? AND stars = 4",id)).Count(c.Context(),boil.GetContextDB())
	detailedResponse.Stars.FourStars = int(fourStars)
	threeStars, _ := models.Reviews(qm.Where("\"productID\" = ? AND stars = 3",id)).Count(c.Context(),boil.GetContextDB())
	detailedResponse.Stars.ThreeStars = int(threeStars)
	twoStars, _ := models.Reviews(qm.Where("\"productID\" = ? AND stars = 2",id)).Count(c.Context(),boil.GetContextDB())
	detailedResponse.Stars.TwoStars = int(twoStars)
	oneStars, _ := models.Reviews(qm.Where("\"productID\" = ? AND stars = 1",id)).Count(c.Context(),boil.GetContextDB())
	detailedResponse.Stars.OneStars = int(oneStars)

	resp := fiber.Map{
		"status": "Success",
		"data": detailedResponse,
		"totalPage": totalPage,
		"message": "Successfully fetched detailed product!",
	}

	//Saving redis cache for 30 mins
	jsonData, _ := json.Marshal(resp)
	redisdatabase.Client.Set(redisdatabase.Ctx,redisKey,jsonData,30*time.Minute)
	//add this key to the set tracking all cache keys of this product
	redisSetKey := fmt.Sprintf("product:detail:%d:keys",id)
	err = redisdatabase.Client.SAdd(redisdatabase.Ctx,redisSetKey,redisKey).Err()
	if err != nil{
		fmt.Println("Error adding redis key to set: ", err)
	}

	return c.JSON(resp);
}

func reviewDisplay(ctx context.Context,id int, filter string, offset int) (reviews []ReviewResponse,error error, totalPage int){
	queries := []qm.QueryMod{}
	queries = append(queries, qm.Where("\"productID\" = ?",id))

	if filter != "All"{
		star, err := strconv.Atoi(filter);
		if err != nil || star < 1 || star > 5{
			return nil,fmt.Errorf("invalid star filter"), 0;
		}
		queries = append(queries, qm.Where("stars = ?",star))
	}

	//calculating total page
	totalReview, err := models.Reviews(queries...).Count(context.Background(),boil.GetContextDB())
	if err != nil{
		return nil,fmt.Errorf(err.Error()),0;
	}
	totalPage = int(math.Ceil(float64(totalReview)/3));

	//loading necessary references
	queries = append(queries, 
		qm.Load(models.ReviewRels.AccountIDAccount),
		qm.Load(models.ReviewRels.ProductIDProduct),
		qm.Load(models.ReviewRels.ReviewIDReviewImages),
		qm.Load(models.ReviewRels.ReviewIDReplies),
		qm.Offset(offset),
		qm.Limit(3),
		qm.OrderBy("\"reviewID\""),
	)

	review, err := models.Reviews(queries...).All(ctx,boil.GetContextDB());
	if err != nil{
		return nil,err,totalPage
	}
	
	reviewResult := make([]ReviewResponse,len(review));
	for i, r := range review{
		reviewResult[i] = ReviewResponse{
			Review: *r,
			ReviewAccount: *r.R.AccountIDAccount,
			ReviewProduct: *r.R.ProductIDProduct,
			ReviewImages: r.R.ReviewIDReviewImages,
			ReviewReply: r.R.ReviewIDReplies,
		}
	}
	return reviewResult,nil,totalPage;
}

func GetSimilar(c *fiber.Ctx) error{
	productID := c.Query("id","");
	if productID == ""{
		return service.SendError(c,404,"Did not receive ID!");
	}

	typeID := c.Query("typeID","");
	if typeID == ""{
		return service.SendError(c,404,"Did not receive typeID!");
	}

	//Fetching typeName from typeID
	similars,err := models.Products(qm.Where("\"productID\" != ? AND \"productTypeID\" = ?",productID,typeID)).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,"ID not found!");
	}

	resp := fiber.Map{
		"status": "Success",
		"data": similars,
		"message": "Successfully fetched similar products",
	}

	return c.JSON(resp);
}