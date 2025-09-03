package dto

import "GoodFood-BE/models"

//ProductCards struct represents 2 pieces of info in AdminProduct.tsx according to the UI design
type ProductCards struct{
	TotalProduct int `boil:"totalproduct"`
	TotalInactive int `boil:"totalinactive"`
}

//GetFourStruct represents 4 products with types to display at Home.tsx
type GetFourStruct struct{
	models.Product
	ProductType *models.ProductType
}

//PredictResult struct represents the returned results from microservice image recognition
type PredictResult struct{
	Message string `json:"message"`
	ProductName string `json:"productName"`
	Accuracy float64 `json:"accuracy"`
}

//Star struct represents the number of reviews for each rating level.
type Star struct{
	FiveStars int `json:"fiveStars"`
	FourStars int `json:"fourStars"`
	ThreeStars int `json:"threeStars"`
	TwoStars int `json:"twoStars"`
	OneStars int `json:"oneStars"`
}

//ProductDetailResponse struct represents the detailed info of a product. Used to display info at ProductDetail.tsx
type ProductDetailResponse struct{
	models.Product `json:"product"`
	ProductImages models.ProductImageSlice `json:"productImages"`
	FiveStarsReview []ReviewResponse `json:"review"`
	Stars Star `json:"stars"`
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