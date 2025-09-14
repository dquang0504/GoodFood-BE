package dto

type DashboardValues struct {
	TotalProductSold int     `boil:"total_product_sold" json:"totalProductSold"`
	TotalIncome      float64 `boil:"total_income" json:"totalIncome"`
	TotalUser        int     `boil:"total_user" json:"totalUser"`
	TotalInvoice     int     `boil:"total_invoice" json:"totalInvoice"`
}

type MonthlyIncome struct {
	Month       int     `boil:"month" json:"month"`
	TotalIncome float64 `boil:"total_income" json:"totalIncome"`
}

type PieChart struct {
	Label string  `boil:"label" json:"label"`
	Value float64 `boil:"value" json:"value"`
}

type BarChart struct {
	Month int     `boil:"month" json:"month"`
	Value float64 `boil:"value" json:"value"`
}