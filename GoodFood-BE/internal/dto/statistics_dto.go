package dto

// StatisticsResponse represents the aggregated sales and revenue grouped by product type.
type StatisticsResponse struct{
	ProductType string `boil:"product_type" json:"productType"`
	TotalSale int64 `boil:"total_sale" json:"totalSale"`
	TotalRevenue float64 `boil:"total_revenue" json:"totalRevenue"`
}