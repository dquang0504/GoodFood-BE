package handlers

import (
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