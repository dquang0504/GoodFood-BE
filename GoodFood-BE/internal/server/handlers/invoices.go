package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type InvoicePayload struct{
	Invoice models.Invoice `json:"invoice"`
	InvoiceDetails []models.InvoiceDetail `json:"invoiceDetails"`
}

func InvoicePay(c *fiber.Ctx) error{
	var payload InvoicePayload
	if err := c.BodyParser(&payload); err != nil{
		return service.SendError(c,401,"Invalid body details: " + err.Error());
	}

	//insert invoice first
	if err := payload.Invoice.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}

	//insert invoice details
	for _,detail := range payload.InvoiceDetails{
		detail.InvoiceID = payload.Invoice.InvoiceID
		if err := detail.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
			return service.SendError(c, 500, err.Error())
		}
	}
	
	resp := fiber.Map{
		"status": "Success",
		"data": payload.InvoiceDetails,
		"message": "Successfully created new invoice",
	}

	return c.JSON(resp);

}