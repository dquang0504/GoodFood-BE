package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"math"
	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type InvoiceCards struct{
	TotalInvoice int `boil:"total"`
	TotalCanceled int `boil:"canceled"`
}

type InvoiceResponse struct{
	models.Invoice
	InvoiceDetails models.InvoiceDetailSlice `json:"invoiceDetails"`
	InvoiceStatus *models.InvoiceStatus `json:"invoiceStatus"`
	Product *models.Product `json:"product"`
}

func GetAdminInvoice(c *fiber.Ctx) error{
	var cards InvoiceCards
	err := queries.Raw(`
		SELECT COALESCE(COUNT("invoiceID"),0) AS total,
		COUNT(CASE WHEN "invoiceStatusID" = 6 THEN 1 END) AS canceled
		FROM invoice
	`).Bind(c.Context(),boil.GetContextDB(),&cards)
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	//Lấy về số trang
	page := c.QueryInt("page",0);
	if page == 0{
		return service.SendError(c,401,"Did not receive page");
	}
	//Lấy về sort và search
	sort := c.Query("sort","");
	search := c.Query("search","")

	//calculating offset
	offset := (page-1)*6;

	//creating query mod
	queryMods := []qm.QueryMod{
		qm.Limit(6),
		qm.Offset(offset),
		qm.Load(models.InvoiceRels.InvoiceIDInvoiceDetails),
		qm.Load(models.InvoiceRels.InvoiceStatusIDInvoiceStatus),
		qm.OrderBy("\"invoiceID\" DESC"),
	}

	//working on sort and search logic
	if search != ""{
		switch sort{
			case "Mã hóa đơn":
				queryMods = append(queryMods, qm.Where("CAST(\"invoiceID\" AS TEXT) ILIKE ?", "%"+search+"%"))
			case "Tên khách hàng":
				queryMods = append(queryMods, qm.Where("\"receiveName\" ILIKE ?","%"+search+"%"))
		}
	}


	invoices,err := models.Invoices(
		queryMods...
	).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	totalPage := int(math.Ceil(float64(len(invoices)) / float64(6)))


	// Converting data into custom struct that has invoice, invoice details and its statuses
	response := make([]InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		response[i] = InvoiceResponse{
			Invoice: *invoice,
			InvoiceStatus: invoice.R.InvoiceStatusIDInvoiceStatus,
			InvoiceDetails: invoice.R.InvoiceIDInvoiceDetails,
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"cards": cards,
		"totalPage": totalPage,
		"message": "Successfully fetched invoice values",
	}

	return c.JSON(resp);
}