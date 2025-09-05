package dto

import (
	"GoodFood-BE/models"

	"github.com/aarondl/null/v8"
)

//InvoiceCards struct represents 2 pieces of info in AdminOrder.tsx according to UI design
type InvoiceCards struct{
	TotalInvoice int `boil:"total"`
	TotalCanceled int `boil:"canceled"`
}

//InvoiceResponse struct represents invoice data with related entities for frontend readability
type InvoiceResponse struct{
	models.Invoice
	InvoiceStatus *models.InvoiceStatus `json:"invoiceStatus"`
	Product *models.Product `json:"product"`
}

//UpdateInvoiceStruct struct represents invoice metrics used for updating invoices.
type UpdateInvoiceStruct struct{
	StatusName string `json:"statusName"`
	CancelReason null.String `json:"cancelReason"`
}

// InvoiceDetailResponse is a DTO for joining invoice_detail, product, and invoice info.
// Avoid multiple nested objects.
type InvoiceDetailResponse struct{
	InvoiceDetailID int     `boil:"invoiceDetailID"`
	InvoiceID       int     `boil:"invoiceID"`
	ProductID       int     `boil:"productID"`
	Price           float64 `boil:"price"`
	Quantity        int     `boil:"quantity"`

	ProductName string `boil:"food"`
	ReceiveName string `boil:"name"`
	ReceivePhone string `boil:"phone"`
	ReceiveAddress string `boil:"address"`
}