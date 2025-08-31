package handlers

import (
	"GoodFood-BE/internal/dto"
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func GetReviewData(c *fiber.Ctx) error{
	invoiceID := c.QueryInt("invoiceID",0);
	productID := c.QueryInt("productID",0);
	if invoiceID == 0 || productID == 0{
		return service.SendError(c,400,"Did not receive invoiceID or productID");
	}

	invoiceDetails, err := models.InvoiceDetails(
		qm.Where("\"invoiceID\" = ?",invoiceID),
		qm.Load(models.InvoiceDetailRels.ProductIDProduct),
		qm.Load(models.InvoiceDetailRels.InvoiceIDInvoice),
	).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	response := InvoiceDetailStruct{}
	for _, detail := range invoiceDetails{
		check := false;
		review, err := models.Reviews(
			qm.Where("\"productID\" = ?",detail.R.ProductIDProduct.ProductID),
		).One(c.Context(),boil.GetContextDB());
		if err == nil && review != nil{
			check = true;	
		}
		response = InvoiceDetailStruct{
			InvoiceID: detail.InvoiceID,
			Image: detail.R.ProductIDProduct.CoverImage,
			Product: *detail.R.ProductIDProduct,
			Quantity: detail.Quantity,
			TotalMoney: float64(detail.Price),
			ShippingFee: float64(detail.R.InvoiceIDInvoice.ShippingFee),
			ReviewCheck: check,
		}
		
	}

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"message": "Successfully fetched product to review!",
	}
	return c.JSON(resp);
}
type ReviewSubmitResponse struct{
	models.Review
	ReviewImages []models.ReviewImage `json:"reviewImages"`
}
type NSFWScores struct{
	Unsafe float64 `json:"unsafe"`
	Porn float64 `json:"porn"`
	Sexy float64 `json:"sexy"`
}
type ImageDetectionResult struct{
	Image string `json:"image"`
	NSFW bool `json:"nsfw"`
	NSFWScores NSFWScores `json:"nsfw_scores"`
	Violent bool `json:"violent"`
	ViolentLabel string `json:"violent_label"`
}
type ReviewContentDetection struct{
	Label string `json:"label"`
	Score float64 `json:"score"`
	Images []ImageDetectionResult `json:"images"`
}
func HandleSubmitReview(c *fiber.Ctx) error{
	//resty
	client := resty.New();

	// 1. Parse phần "review" từ multipart
	reviewJson := c.FormValue("review")
	if reviewJson == "" {
		return service.SendError(c, 400, "Missing review data")
	}

	var body ReviewSubmitResponse
	if err := json.Unmarshal([]byte(reviewJson), &body); err != nil {
		return service.SendError(c, 400, "Invalid review JSON: "+err.Error())
	}

	// 2. Lấy tất cả file "images"
	form, err := c.MultipartForm()
	if err != nil {
		return service.SendError(c, 400, "Error parsing multipart: "+err.Error())
	}
	files := form.File["reviewImages"]

	// 3. Đọc binary các file ảnh để gửi cho gRPC Flask
	var imageBinaries = make(map[string][]byte)
	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			return service.SendError(c, 500, "Failed to open image: "+err.Error())
		}
		defer f.Close()

		buf, err := io.ReadAll(f)
		if err != nil {
			return service.SendError(c, 500, "Failed to read image: "+err.Error())
		}
		imageBinaries[file.Filename] = buf
	}

	// 4. Gửi ảnh binary & comment tới Flask để kiểm duyệt
	payload := map[string]interface{}{
		"review": body.Comment,
		"images": imageBinaries,
	}

	var result ReviewContentDetection
	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(&result).
		Post("http://192.168.1.10:5000/reviewLabel")

	if err != nil {
		return service.SendError(c, 500, "Failed to call gRPC Flask: "+err.Error())
	}

	if result.Label == "toxic" {
		return service.SendError(c, 400, "⚠️ Hate Speech Detected: Please edit your comment.")
	}

	//NSFW and violence detection alerts
	for _, img := range result.Images {
		fmt.Println(img);
		if img.NSFW {
			return service.SendError(c, 400, "⚠️ NSFW Content Detected. Please remove or replace the image: "+img.Image)
		}

		if img.Violent{
			return service.SendError(c, 400, "⚠️ Violence Detected. Please remove or replace the image: "+img.Image)
		}
	}

	//insert new review
	err = body.Insert(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//also clearing product cache after insertion
	err = utils.ClearProductCache(body.ProductID)
	if err != nil{
		fmt.Println("Error clearing product cache: ",err)
	}

	//insert images into firebase storage
	uploadedURLs, err := utils.UploadFirebaseImages(imageBinaries,c.Context());
	if err != nil{
		return service.SendError(c,500, err.Error())
	}

	//insert its corresponding review images
	var reviewImages = body.ReviewImages
	for _, url := range uploadedURLs{
		reviewImages = append(reviewImages, models.ReviewImage{
			ReviewID: body.ReviewID,
			ImageName: url,
		})
	}

	for _, img := range reviewImages{
		err = img.Insert(c.Context(),boil.GetContextDB(),boil.Infer());
		if err != nil{
			return service.SendError(c,500,err.Error());
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": body,
		"result": result,
		"message": "Successfully submitted product review!",
	}
	return c.JSON(resp);
}

func GetReviewDetail(c *fiber.Ctx) error{
	reviewID := c.QueryInt("reviewID",0);
	if reviewID == 0{
		return service.SendError(c,400, "Did not receive reviewID!");
	}

	review, err := models.Reviews(
		qm.Where("\"reviewID\" = ?",reviewID),
		qm.Load(models.ReviewRels.AccountIDAccount),
		qm.Load(models.ReviewRels.ProductIDProduct),
		qm.Load(models.ReviewRels.ReviewIDReviewImages),
		qm.Load(models.ReviewRels.ReviewIDReplies),
		qm.Load(models.ReviewRels.InvoiceIDInvoice),
	).One(c.Context(),boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	
	invoiceDetails, err := models.InvoiceDetails(qm.Where("\"invoiceID\" = ? AND \"productID\" = ?",review.InvoiceID,review.ProductID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		fmt.Println(review.InvoiceID);
		fmt.Println(review.ProductID);
		return service.SendError(c,500,err.Error());
	}

	response := dto.ReviewResponse{
		Review: *review,
		ReviewAccount: *review.R.AccountIDAccount,
		ReviewProduct: *review.R.ProductIDProduct,
		ReviewImages: review.R.ReviewIDReviewImages,
		ReviewReply: review.R.ReviewIDReplies,
		ReviewInvoice: *review.R.InvoiceIDInvoice,
	}

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"detail": invoiceDetails,
		"message": "Successfully fetched review detail!",
	}
	return c.JSON(resp);
}

func HandleUpdateReview(c *fiber.Ctx) error{
	//resty
	client := resty.New();

	// 1. Parse phần "review" từ multipart
	reviewJson := c.FormValue("review")
	if reviewJson == "" {
		return service.SendError(c, 400, "Missing review data")
	}


	reviewID := c.QueryInt("reviewID",0);
	if reviewID == 0{
		return service.SendError(c,400,"Did not receive reviewID");
	}
	var body ReviewSubmitResponse
	if err := json.Unmarshal([]byte(reviewJson), &body); err != nil {
		return service.SendError(c, 400, "Invalid review JSON: "+err.Error())
	}

	// 2. Lấy tất cả file "images"
	form, err := c.MultipartForm()
	if err != nil {
		return service.SendError(c, 400, "Error parsing multipart: "+err.Error())
	}
	files := form.File["reviewImages"]

	// 3. Đọc binary các file ảnh để gửi cho gRPC Flask
	var imageBinaries = make(map[string][]byte)
	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			return service.SendError(c, 500, "Failed to open image: "+err.Error())
		}
		defer f.Close()

		buf, err := io.ReadAll(f)
		if err != nil {
			return service.SendError(c, 500, "Failed to read image: "+err.Error())
		}
		imageBinaries[file.Filename] = buf
		fmt.Println(imageBinaries);
	}

	// 4. Gửi ảnh binary & comment tới Flask để kiểm duyệt
	payload := map[string]interface{}{
		"review": body.Comment,
		"images": imageBinaries,
	}

	var result ReviewContentDetection
	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(&result).
		Post("http://192.168.1.10:5000/reviewLabel")

	if err != nil {
		return service.SendError(c, 500, "Failed to call gRPC Flask: "+err.Error())
	}

	if result.Label == "toxic" {
		return service.SendError(c, 400, "⚠️ Hate Speech Detected: Please edit your comment.")
	}

	//NSFW and violence detection alerts
	for _, img := range result.Images {
		fmt.Println(img);
		if img.NSFW {
			return service.SendError(c, 400, "⚠️ NSFW Content Detected. Please remove or replace the image: "+img.Image)
		}

		if img.Violent{
			return service.SendError(c, 400, "⚠️ Violence Detected. Please remove or replace the image: "+img.Image)
		}
	}

	//insert images into firebase storage
	var uploadedURLs map[string]string
	if(len(imageBinaries) > 0){
		uploadedURLs, err = utils.UploadFirebaseImages(imageBinaries,c.Context());
		if err != nil{
			return service.SendError(c,500, err.Error())
		}
	}

	//fetching the referred review
	review, err := models.Reviews(qm.Where("\"reviewID\" = ?",reviewID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//update the review
	review.Comment = body.Comment
	review.Stars = body.Stars

	_,err = review.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//insert its corresponding review images and maybe delete previous ones
	var reviewImages = body.ReviewImages
	//if an image is uploaded, delete the previous ones
	if(len(uploadedURLs) > 0){
		for _, url := range uploadedURLs{
			reviewImages = append(reviewImages, models.ReviewImage{
				ReviewID: body.ReviewID,
				ImageName: url,
			})
		}

		getReviewImgs, err := models.ReviewImages(qm.Where("\"reviewID\" = ?",body.ReviewID)).All(c.Context(),boil.GetContextDB());
		if err != nil{
			return service.SendError(c,500,err.Error());
		}

		//deleting previous ones on firebase storage
		for _, img := range getReviewImgs{
			err := utils.DeleteFirebaseImage(img.ImageName,context.Background());
			if err != nil{
				return service.SendError(c,500,err.Error());
			}
		}

		//deleting previous ones in DB
		for _, img := range getReviewImgs{
			_,err = img.Delete(context.Background(),boil.GetContextDB());
			if err != nil{
				return service.SendError(c,500,err.Error());
			}
		}

		//inserting new ones
		for _, img := range reviewImages{
			err = img.Insert(c.Context(),boil.GetContextDB(),boil.Infer());
			if err != nil{
				return service.SendError(c,500,err.Error());
			}
		}
	}

	// Clear all Redis keys related to this product's review filters
	pattern := fmt.Sprintf("product:detail:%d:filter=*:page=*", body.ProductID)
	iter := redisdatabase.Client.Scan(redisdatabase.Ctx, 0, pattern, 0).Iterator()
	for iter.Next(redisdatabase.Ctx) {
		key := iter.Val()
		err := redisdatabase.Client.Del(redisdatabase.Ctx, key).Err()
		if err != nil {
			fmt.Printf("Failed to delete key %s: %v\n", key, err)
		}
	}
	if err := iter.Err(); err != nil {
		fmt.Printf("Error during Redis scan: %v\n", err)
	}

	resp := fiber.Map{
		"status": "Success",
		"data": body,
		"message": "Successfully updated product review!",
	}
	return c.JSON(resp);
}