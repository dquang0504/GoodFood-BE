package handlers

import (
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"fmt"

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
func HandleSubmitReview(c *fiber.Ctx) error{
	body := ReviewSubmitResponse{}
	err := c.BodyParser(&body);
	if err != nil{
		return service.SendError(c,400,err.Error());
	}
	//insert new review
	err = body.Insert(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	//insert its corresponding review images
	var reviewImages = body.ReviewImages
	for i := range reviewImages{
		reviewImages[i].ReviewID = body.ReviewID
		fmt.Println(reviewImages[i]);
		err = reviewImages[i].Insert(c.Context(),boil.GetContextDB(),boil.Infer());
		if err != nil{
			return service.SendError(c,500,err.Error());
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": body,
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
	fmt.Println("Here lies: ",review.AccountID);
	invoiceDetails, err := models.InvoiceDetails(qm.Where("\"invoiceID\" = ? AND \"productID\" = ?",review.InvoiceID,review.ProductID)).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	response := ReviewResponse{
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
	reviewID := c.QueryInt("reviewID",0);
	if reviewID == 0{
		return service.SendError(c,400,"Did not receive reviewID");
	}
	body := ReviewSubmitResponse{}
	err := c.BodyParser(&body);
	if err != nil{
		return service.SendError(c,400,err.Error());
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
	if len(reviewImages) > 0{
		//checking if new images are uploaded then delete every image of the review
		deleteImgs,err := models.ReviewImages(qm.Where("\"reviewID\" = ?",body.ReviewID)).All(c.Context(),boil.GetContextDB());
		if err != nil{
			return service.SendError(c,500,err.Error());
		}
		_,err = deleteImgs.DeleteAll(c.Context(),boil.GetContextDB())
		if err != nil{
			return service.SendError(c,500,err.Error());
		}
		//iterate through images and start inserting one by one image
		for i := range reviewImages{
			reviewImages[i].ReviewID = body.ReviewID;
			if err := reviewImages[i].Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
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