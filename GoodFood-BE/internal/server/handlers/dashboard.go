package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/gofiber/fiber/v2"
)

//GetDashboard returns the values for the 4 info cards on Dashboard panel.
func GetDashboard(c *fiber.Ctx) error{
	var values dto.DashboardValues
	queriesList := []struct{
		sql string
	}{
		{`SELECT COALESCE(SUM(quantity),0) AS total_product_sold FROM invoice_detail`}, //total product sold
		{`SELECT COALESCE(SUM("totalPrice"),0) AS total_income FROM invoice`}, //total income
		{`SELECT COALESCE(COUNT("username"),0) AS total_user FROM account`}, //total user
		{`SELECT COALESCE(COUNT("invoiceID"),0) AS total_invoice FROM invoice`}, //total invoice
	}
	
	for _,q := range queriesList{
		if err := queryAndBind(c,q.sql,&values); err != nil{
			return service.SendError(c,500,err.Error());
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": values,
		"message": "Successfully fetched dashboard values",
	}

	return c.JSON(resp);
}

//GetLineChart fetches the necessary data to paint a line chart.
func GetLineChart(c *fiber.Ctx) error{
	var incomes []dto.MonthlyIncome
	err := queries.Raw(`
		SELECT EXTRACT(MONTH FROM "createdAt") AS month,
		COALESCE(SUM("totalPrice"),0) AS total_income
		FROM invoice WHERE status = false
		GROUP BY month ORDER BY month
	`).Bind(c.Context(),boil.GetContextDB(),&incomes)
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	resp := fiber.Map{
		"status": "Success",
		"data": incomes,
		"message": "Successfully fetched dashboard values",
	}

	return c.JSON(resp)
}

//GetLineChart fetches the necessary data to paint a line chart.
func GetPieChart(c *fiber.Ctx) error{
	var pieChart []dto.PieChart
	err := queries.Raw(`
		SELECT "typeName" as label, COALESCE(SUM(invoice_detail.quantity),0) as value
		FROM product_type INNER JOIN product
		ON product_type."productTypeID" = product."productTypeID" INNER JOIN invoice_detail
		ON invoice_detail."productID" = product."productID"
		GROUP BY "typeName"
	`).Bind(c.Context(),boil.GetContextDB(),&pieChart)
	if err != nil {
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": pieChart,
		"message": "Successfully fetched dashboard values",
	}

	return c.JSON(resp);
}

//GetBarChart fetches the necessary data to paint a line chart.
func GetBarChart(c *fiber.Ctx) error{
	var barChart []dto.BarChart

	err := queries.Raw(`
		SELECT EXTRACT(MONTH FROM "createdAt") AS month,
		COALESCE(SUM("totalPrice"),0) as value
		FROM invoice WHERE status = false
		GROUP BY month ORDER BY month
	`).Bind(c.Context(),boil.GetContextDB(),&barChart)
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": barChart,
		"message": "Successfully fetched dashboard values",
	}

	return c.JSON(resp);
}

func queryAndBind(c *fiber.Ctx, sql string, dest interface{}) error {
	return queries.Raw(sql).Bind(c.Context(), boil.GetContextDB(), dest)
}