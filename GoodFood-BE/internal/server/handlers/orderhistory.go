package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type InvoiceList struct{
	InvoiceID int `boil:"invoice_id" json:"invoiceID"`
	TotalProducts int `boil:"total_products" json:"totalProducts"`
	Address string `boil:"address" json:"address"`
	Status bool `boil:"status" json:"status"`
	TotalMoney float64 `boil:"total_money" json:"totalMoney"`
	CancelReason string `boil:"cancel_reason" json:"cancelReason"`
}

func GetOrderHistory(c *fiber.Ctx) error{
	tab := c.Query("tab","Đã đặt hàng");

	invoiceList := []InvoiceList{}
	err := queries.Raw(`
		SELECT invoice_detail."invoiceID" as invoice_id, COALESCE(COUNT(invoice_detail."productID"),0) as total_products,
		invoice."receiveAddress" as address, invoice.status as status, invoice."totalPrice" as total_money,
		invoice."cancelReason" as cancel_reason
		FROM invoice INNER JOIN invoice_detail
		ON invoice."invoiceID" = invoice_detail."invoiceID"
		INNER JOIN invoice_status ON invoice."invoiceStatusID" = invoice_status."invoiceStatusID"
		WHERE invoice_status."statusName" = $1
		GROUP BY invoice_detail."invoiceID", invoice."receiveAddress", invoice.status, invoice."totalPrice", invoice."cancelReason"
	`,tab).Bind(c.Context(),boil.GetContextDB(),&invoiceList);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": invoiceList,
		"message": "Successfully fetched invoice list!",
	}

	return c.JSON(resp);

}

func CancelOrder(c *fiber.Ctx) error{
	var invoice InvoiceList
	invoiceID := c.QueryInt("invoiceID",0);
	if invoiceID == 0{
		return service.SendError(c,400, "Did not receive invoiceID!");
	}
	if err := c.BodyParser(&invoice); err != nil{
		return service.SendError(c,400,err.Error());
	}

	toUpdate, err := models.Invoices(qm.Where("\"invoiceID\" = ?",invoiceID)).One(c.Context(),boil.GetContextDB());
	if err != nil {
		return service.SendError(c,500,err.Error());
	}
	
	toUpdate.InvoiceStatusID = 6
	toUpdate.CancelReason = null.StringFrom(invoice.CancelReason)
	_ ,err = toUpdate.Update(c.Context(),boil.GetContextDB(),boil.Infer());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	resp := fiber.Map{
		"status": "Success",
		"data": toUpdate,
		"message": "Successfully canceled the order!",
	}

	return c.JSON(resp);

}

type InvoiceDetailStruct struct{
	InvoiceID int `boil:"invoice_id" json:"invoiceID"`
	Image string `boil:"image" json:"image"`
	Product models.Product `boil:"product" json:"product"`
	Quantity int `boil:"quantity" json:"quantity"`
	TotalMoney float64 `boil:"total_money" json:"totalMoney"`
	ShippingFee float64 `boil:"shipping_fee" json:"shippingFee"`
	ReviewCheck bool `json:"reviewCheck"`
}

func GetOrderHistoryDetail(c *fiber.Ctx) error{
	invoiceID := c.QueryInt("invoiceID",0);
	if invoiceID == 0{
		return service.SendError(c,400,"Did not receive invoiceID!");
	}

	invoiceDetails, err := models.InvoiceDetails(
		qm.Where("\"invoiceID\" = ?",invoiceID),
		qm.Load(models.InvoiceDetailRels.ProductIDProduct),
		qm.Load(models.InvoiceDetailRels.InvoiceIDInvoice),
	).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	response := make([]InvoiceDetailStruct,len(invoiceDetails));
	for i, detail := range invoiceDetails{
		check := false;
		review, err := models.Reviews(
			qm.Where("\"productID\" = ?",detail.R.ProductIDProduct.ProductID),
		).One(c.Context(),boil.GetContextDB());
		if err == nil && review != nil{
			check = true;	
		}
		response[i] = InvoiceDetailStruct{
			InvoiceID: detail.InvoiceID,
			Image: detail.R.ProductIDProduct.CoverImage,
			Product: *detail.R.ProductIDProduct,
			Quantity: detail.Quantity,
			TotalMoney: float64(detail.Price),
			ShippingFee: float64(detail.R.InvoiceIDInvoice.ShippingFee),
			ReviewCheck: check,
		}
		
	}



	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"message": "Successfully fetched receipt details!",
	}

	return c.JSON(resp);
}