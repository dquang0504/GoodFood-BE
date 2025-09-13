package dto

import "GoodFood-BE/models"

//CartDetailResponse struct represents the api response of Cart module.
type CartDetailResponse struct{
	models.CartDetail
	Product *models.Product `json:"product"`
}