package utils

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/models"
)

func BuildCartResponse(carts []*models.CartDetail) []dto.CartDetailResponse{
	response := make([]dto.CartDetailResponse, len(carts))
	for i, cart := range carts{
		response[i] = dto.CartDetailResponse{
			CartDetail: *cart,
			Product: cart.R.ProductIDProduct,
		}
	}
	return response;
}