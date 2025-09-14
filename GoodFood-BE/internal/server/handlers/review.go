package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

//GetReviewData returns a list of reviews of a specific product.
func GetReviewData(c *fiber.Ctx) error{
	//Fetch query params
	invoiceID := c.QueryInt("invoiceID",0);
	productID := c.QueryInt("productID",0);
	if invoiceID == 0 || productID == 0{
		return service.SendError(c,400,"Did not receive invoiceID or productID");
	}

	//Build response
	invoiceDetails, err := models.InvoiceDetails(
		qm.Where("\"invoiceID\" = ?",invoiceID),
		qm.Load(models.InvoiceDetailRels.ProductIDProduct),
		qm.Load(models.InvoiceDetailRels.InvoiceIDInvoice),
	).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	response := dto.InvoiceDetailStruct{}
	for _, detail := range invoiceDetails{
		reviewExist, _ := models.Reviews(
			qm.Where("\"productID\" = ?",detail.R.ProductIDProduct.ProductID),
		).Exists(c.Context(),boil.GetContextDB());

		response = dto.InvoiceDetailStruct{
			InvoiceID: detail.InvoiceID,
			Image: detail.R.ProductIDProduct.CoverImage,
			Product: *detail.R.ProductIDProduct,
			Quantity: detail.Quantity,
			TotalMoney: float64(detail.Price),
			ShippingFee: float64(detail.R.InvoiceIDInvoice.ShippingFee),
			ReviewCheck: reviewExist,
		}
		
	}

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"message": "Successfully fetched product to review!",
	}
	return c.JSON(resp);
}

func HandleSubmitReview(c *fiber.Ctx) error{
	//resty
	client := resty.New();

	// 1. Parse "review" from multipart
	reviewJson := c.FormValue("review")
	if reviewJson == "" {
		return service.SendError(c, 400, "Missing review data")
	}

	var body dto.ReviewSubmitRequest
	if err := json.Unmarshal([]byte(reviewJson), &body); err != nil {
		return service.SendError(c, 400, "Invalid review JSON: "+err.Error())
	}

	// 2. Fetch all files from "images"
	form, err := c.MultipartForm()
	if err != nil {
		return service.SendError(c, 400, "Error parsing multipart: "+err.Error())
	}
	files := form.File["reviewImages"]

	// 3. Read images binary and send to microservice Flask
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

	// 4. Send binary images & comment to Flask to moderate
	payload := map[string]interface{}{
		"review": body.Comment,
		"images": imageBinaries,
	}

	var result dto.ReviewContentDetection
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
	redisSetKey := fmt.Sprintf("product:detail:%d:keys",body.ProductID)
	utils.ClearCache(redisSetKey);

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

//GetReviewDetail returns all the details of the specified review
func GetReviewDetail(c *fiber.Ctx) error{
	//Fetch query param
	reviewID := c.QueryInt("reviewID",0);
	if reviewID == 0{
		return service.SendError(c,400, "Did not receive reviewID!");
	}

	//Build response
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

//HandleUpdateReview updates the specified review
func HandleUpdateReview(c *fiber.Ctx) error{
	//resty
	client := resty.New();

	//Fetch query param
	reviewID := c.QueryInt("reviewID",0);
	if reviewID == 0{
		return service.SendError(c,400,"Did not receive reviewID");
	}

	// 1. Parse "review" from multipart
	reviewJson := c.FormValue("review")
	if reviewJson == "" {
		return service.SendError(c, 400, "Missing review data")
	}
	var body dto.ReviewSubmitRequest
	if err := json.Unmarshal([]byte(reviewJson), &body); err != nil {
		return service.SendError(c, 400, "Invalid review JSON: "+err.Error())
	}

	// 2. Fetch all files from "images"
	form, err := c.MultipartForm()
	if err != nil {
		return service.SendError(c, 400, "Error parsing multipart: "+err.Error())
	}
	files := form.File["reviewImages"]

	// 3. Read images binary to send to Flask microservice
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

	// 4. Send binary images & comment to Flask for moderation
	payload := map[string]interface{}{
		"review": body.Comment,
		"images": imageBinaries,
	}

	var result dto.ReviewContentDetection
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
	utils.ClearCache(pattern);

	resp := fiber.Map{
		"status": "Success",
		"data": body,
		"message": "Successfully updated product review!",
	}
	return c.JSON(resp);
}