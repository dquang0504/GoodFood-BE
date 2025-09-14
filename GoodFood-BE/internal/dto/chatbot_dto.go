package dto

type PromptRequest struct {
	Prompt string `json:"prompt"`
}

type BestSellingProductResponse struct {
	ProductID         int    `boil:"productID" json:"productID"`
	ProductName       string `boil:"productName" json:"productName"`
	TotalQuantitySold int    `boil:"total_quantity_sold" json:"totalQuantitySold"`
	ProductImage string `boil:"product_img" json:"productImage"`
}