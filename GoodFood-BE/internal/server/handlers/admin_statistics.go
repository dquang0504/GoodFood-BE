package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/gofiber/fiber/v2"
)

// GetAdminStatistics provides sales and revenue statistics by product type within a date range.
func GetAdminStatistics(c *fiber.Ctx) error{
	filter := c.Query("filter","");
	if filter == ""{
		return service.SendError(c,400,"Did not receive filter!");
	}

	//Parse and validate date range
	dateFrom,dateTo,err := utils.ParseDateRange(c.Query("ngayFrom",""),c.Query("ngayTo",""));
	if err != nil{
		return service.SendError(c,400,err.Error());
	}

	//Fetch statistics with raw SQL query
	var statistics []dto.StatisticsResponse
	err = queries.Raw(`
		SELECT product_type."typeName" AS product_type, 
				COALESCE(SUM(invoice_detail.quantity),0) AS total_sale,
				COALESCE(SUM(invoice_detail.price * invoice_detail.quantity),0) AS total_revenue
		FROM product_type 
		INNER JOIN product ON product."productTypeID" = product_type."productTypeID"
		INNER JOIN invoice_detail ON product."productID" = invoice_detail."productID"
		INNER JOIN invoice ON invoice."invoiceID" = invoice_detail."invoiceID"
		WHERE invoice."createdAt" BETWEEN $1 AND $2
		GROUP BY product_type."typeName"
	`,dateFrom,dateTo).Bind(c.Context(),boil.GetContextDB(),&statistics)
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