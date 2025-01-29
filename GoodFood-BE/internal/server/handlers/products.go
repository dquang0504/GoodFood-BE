package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"encoding/base64"
	"math"
	"sort"
	"strconv"

	"github.com/gofiber/fiber/v2"
	tflite "github.com/mattn/go-tflite"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	// "golang.org/x/text/number"
)

func GetFour(c *fiber.Ctx) error {
	println("HELLO TÊACHER")
    // Tạo context từ Fiber
	ctx := c.Context()

	// Truy vấn danh sách sản phẩm
	products, err := models.Products(qm.Limit(4)).All(ctx, boil.GetContextDB())
	if err != nil {
		return service.SendError(c,500,"Faield to fetch products")
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
	boil.DebugMode = true
	var totalProduct int64

	//Lấy về số trang
	page,err := strconv.Atoi(c.Query("page","1"));
	if err != nil{
		return service.SendError(c,500,"Error converting pageNum");
	}

	//Lấy về các tham số
	typeName := c.Query("type","");
	search := c.Query("search","");
	//Tính offset
	offset := (page-1)*6;

	//Tạo query mod
	queryMods := []qm.QueryMod{
		qm.Load(models.ProductRels.ProductTypeIDProductType),
	}

	//Nếu có typeName thì lấy totalProduct theo typeName
	if typeName != ""{
		productType,err := models.ProductTypes(qm.Where("\"typeName\" = ?",typeName)).One(c.Context(),boil.GetContextDB())
		if err != nil {
			return service.SendError(c, 500, "Product type not found")
		}
		queryMods = append(queryMods, qm.Where( "\"productTypeID\" = ?",productType.ProductTypeID))
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
		return service.SendError(c, 500, "Total product not found")
	}

	//Thêm vào offset và limit để phân trang
	queryMods = append(queryMods, qm.Limit(6), qm.Offset(offset));
	products, err := models.Products(queryMods...).All(c.Context(), boil.GetContextDB())
	
	
	if err != nil {
		println(err.Error())
		return service.SendError(c, 500, "Failed to fetch products by page")
	}

	totalPage := int(math.Ceil(float64(totalProduct) / float64(6)))

	resp := fiber.Map{
		"status": "Success",
		"data": products,
		"totalPage": totalPage,
		"message": "Successfully fetched products by page",
	}

	println(totalPage)

	return c.JSON(resp);
}

var classNames []string = []string{
	"Bánh flan", "Bánh mì ngọt", "Bánh mochi", "Bánh tiramisu",
	"Chè thái", "Cơm bò lúc lắc", "Cơm cá chiên", "Cơm chiên dương châu", "Cơm gà", "Cơm tấm",
	"Cơm thịt kho", "Cơm xá xíu", "Kem dừa", "Kem socola", "Nước ngọt 7up", "Nước ngọt coca-cola",
	"Nước ngọt pepsi", "Nước ngọt sprite", "Nước tăng lực red bull", "Nước tăng lực sting", "Thịt bò hầm tiêu xanh",
	"Thịt heo quay",
}

func ClassifyImage(c *fiber.Ctx) error{
	println("Hello teacher")
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
	if err := interpreter.AllocateTensors(); err == tflite.Error{
		return service.SendError(c,500,"Failed to allocate tensors")
	}

	//Lấy dữ liệu ảnh từ request
	imageBase64 := c.Query("image","");
	if imageBase64 == ""{
		return service.SendError(c,400,"No image found!");
	}

	//Giải mã base64 thành dữ liệu ảnh
	imageBytes, err := base64.StdEncoding.DecodeString(imageBase64);
	if err != nil{
		return service.SendError(c,500,"Failed to decode image base64")
	}

	//Chuyển ảnh thành tensor
	tensorData,err := processImageForModel(imageBytes);
	if err != nil{
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

	//Trả về kết quả
	return c.JSON(fiber.Map{
		"status": "Success",
		"message": "Image classified successfully",
		"data": resultWithConfidence,
	})

}

func processImageForModel(imageBytes []byte) ([]float32, error){
	return make([]float32, 224*224*3),nil
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