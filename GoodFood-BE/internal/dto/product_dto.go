package dto

import "GoodFood-BE/models"

//ProductCards struct represents 2 pieces of info in AdminProduct.tsx according to the UI design
type ProductCards struct{
	TotalProduct int `boil:"totalproduct"`
	TotalInactive int `boil:"totalinactive"`
}

//ProductResponse struct represents product data with related entities for frontend readability
type ProductResponse struct{
	models.Product
	ProductType models.ProductType `json:"productType"`
	ProductImages []models.ProductImage `json:"productImages"`
}

//ProductError defines validation errors for product creation/update
type ProductError struct {
	ErrProductName string `json:"errProductName"`
	ErrPrice       string `json:"errPrice"`
	ErrWeight      string `json:"errWeight"`
	ErrType        string `json:"errType"`
	ErrImages      string `json:"errImages"`
}