package dto

import "GoodFood-BE/models"

const(
	PageSize = 6
)

type ProductTypesResponse struct {
	models.ProductType
	TotalProduct int `boil:"totalproduct"`
}

type ProductTypeError struct {
	ErrTypeName string `json:"errTypeName"`
}