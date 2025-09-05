package handlers

import (
	"GoodFood-BE/internal/service"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/gofiber/fiber/v2"
)

type StatisticsResponse struct{
	ProductType string `boil:"product_type" json:"productType"`
	TotalSale int64 `boil:"total_sale" json:"totalSale"`
	TotalRevenue float64 `boil:"total_revenue" json:"totalRevenue"`
}
func GetAdminStatistics(c *fiber.Ctx) error{
	filter := c.Query("filter","");
	if filter == ""{
		return service.SendError(c,400,"Did not receive filter!");
	}
	//fetching date queries
	ngayFromStr := c.Query("ngayFrom","");
	ngayToStr := c.Query("ngayTo","");
	//parsing dates
	var ngayFrom, ngayTo time.Time
	var errTime error
	if ngayFromStr != ""{
		ngayFrom, errTime = time.Parse("2006-01-02",ngayFromStr);
		if errTime != nil{
			return service.SendError(c,400,"Invalid format for ngayFrom (expect yyyy-mm-dd)")
		}
	}
	if ngayToStr != ""{
		ngayTo, errTime = time.Parse("2006-01-02",ngayToStr);
		if errTime != nil{
			return service.SendError(c,400,"Invalid format for ngayTo (expect yyyy-mm-dd)")
		}
	}

	if(ngayTo.Before(ngayFrom)){
		return service.SendError(c,400,"Date from can't be before date to!")
	}

	var statistics []StatisticsResponse
	err := queries.Raw(`
		SELECT product_type."typeName" AS product_type, 
				COALESCE(SUM(invoice_detail.quantity),0) AS total_sale,
				COALESCE(SUM(invoice_detail.price * invoice_detail.quantity),0) AS total_revenue
		FROM product_type 
		INNER JOIN product ON product."productTypeID" = product_type."productTypeID"
		INNER JOIN invoice_detail ON product."productID" = invoice_detail."productID"
		INNER JOIN invoice ON invoice."invoiceID" = invoice_detail."invoiceID"
		WHERE invoice."paymentDate" BETWEEN $1 AND $2
		GROUP BY product_type."typeName"
	`,ngayFrom,ngayTo).Bind(c.Context(),boil.GetContextDB(),&statistics)
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": statistics,
		"message": "Successfully fetched statistical data!",
	}
	
	return c.JSON(resp);
}