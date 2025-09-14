package dto

import "GoodFood-BE/models"

//InvoiceList struct represents the response body returned to front-end.
type InvoiceList struct{
	InvoiceID int `boil:"invoice_id" json:"invoiceID"`
	TotalProducts int `boil:"total_products" json:"totalProducts"`
	Address string `boil:"address" json:"address"`
	Status bool `boil:"status" json:"status"`
	TotalMoney float64 `boil:"total_money" json:"totalMoney"`
	CancelReason string `boil:"cancel_reason" json:"cancelReason"`
}

//InvoiceDtailStruct struct represents the response body returned to front-end when asked for the invoice details.
type InvoiceDetailStruct struct{
	InvoiceID int `boil:"invoice_id" json:"invoiceID"`
	Image string `boil:"image" json:"image"`
	Product models.Product `boil:"product" json:"product"`
	Quantity int `boil:"quantity" json:"quantity"`
	TotalMoney float64 `boil:"total_money" json:"totalMoney"`
	ShippingFee float64 `boil:"shipping_fee" json:"shippingFee"`
	ReviewCheck bool `json:"reviewCheck"`
}