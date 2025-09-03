package handlers

import (
	"GoodFood-BE/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
)

type DashboardValues struct{
	TotalProductSold int `boil:"total_product_sold"`
	TotalIncome float64 `boil:"total_income"`
	TotalUser int `boil:"total_user"`
	TotalInvoice int `boil:"total_invoice"`
}
func GetDashboard(c *fiber.Ctx) error{
	var values DashboardValues

	//total product sold
	err1 := queries.Raw(`
		SELECT COALESCE(SUM(quantity),0) AS total_product_sold
			FROM invoice_detail
	`).Bind(c.Context(),boil.GetContextDB(), &values)
	if err1 != nil{
		return service.SendError(c,500,err1.Error())
	}

	//total income
	err2 := queries.Raw(`
		SELECT COALESCE(SUM("totalPrice"),0) AS total_income
		FROM invoice
	`).Bind(c.Context(),boil.GetContextDB(), &values)
	if err2 != nil{
		return service.SendError(c,500,err2.Error())
	}

	//total user
	err3 := queries.Raw(`
		SELECT COALESCE(COUNT("username"),0) AS total_user
		FROM account
	`).Bind(c.Context(),boil.GetContextDB(),&values)
	if err3 != nil{
		return service.SendError(c,500,err3.Error())
	}

	//total invoice
	err4 := queries.Raw(`
		SELECT COALESCE(COUNT("invoiceID"),0) AS total_invoice
		FROM invoice
	`).Bind(c.Context(),boil.GetContextDB(),&values)
	if err4 != nil{
		return service.SendError(c,500,err4.Error())
	}

	resp := fiber.Map{
		"status": "Success",
		"data": values,
		"message": "Successfully fetched dashboard values",
	}

	return c.JSON(resp);
}

type MonthlyIncome struct{
	Month int `boil:"month"`
	TotalIncome float64 `boil:"total_income"`
}
func GetLineChart(c *fiber.Ctx) error{
	var incomes []MonthlyIncome
	err := queries.Raw(`
		SELECT EXTRACT(MONTH FROM "paymentDate") AS month,
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

type PieChart struct{
	Label string `boil:"label"`
	Value float64 `boil:"value"`
}
func GetPieChart(c *fiber.Ctx) error{
	var pieChart []PieChart
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

type BarChart struct{
	Month int `boil:"month"`
	Value float64 `boil:"value"`
}

func GetBarChart(c *fiber.Ctx) error{
	var barChart []BarChart

	err := queries.Raw(`
		SELECT EXTRACT(MONTH FROM "paymentDate") AS month,
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