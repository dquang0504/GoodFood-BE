package handlers

import (
	"GoodFood-BE/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
)

type InvoiceList struct{
	InvoiceID int `boil:"invoice_id" json:"invoiceID"`
	TotalProducts int `boil:"total_products" json:"totalProducts"`
	Address string `boil:"address" json:"address"`
	Status bool `boil:"status" json:"status"`
	TotalMoney float64 `boil:"total_money" json:"totalMoney"`
}

func GetOrderHistory(c *fiber.Ctx) error{
	invoiceList := []InvoiceList{}
	err := queries.Raw(`
		SELECT invoice_detail."invoiceID" as invoice_id, COALESCE(COUNT(invoice_detail."productID"),0) as total_products,
		invoice."receiveAddress" as address, invoice.status as status, invoice."totalPrice" as total_money
		FROM invoice INNER JOIN invoice_detail
		ON invoice."invoiceID" = invoice_detail."invoiceID"
		GROUP BY invoice_detail."invoiceID", invoice."receiveAddress", invoice.status, invoice."totalPrice"
	`).Bind(c.Context(),boil.GetContextDB(),&invoiceList);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": invoiceList,
		"message": "Successfully fetched invoice list!",
	}

	return c.JSON(resp);

}