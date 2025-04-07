package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"

	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/null/v8"
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

	//Fetching page number
	page := c.QueryInt("page",0);
	if page == 0{
		return service.SendError(c,401,"Did not receive page");
	}
	//Fetching sort and search
	sort := c.Query("sort","");
	search := c.Query("search","")
	//Fetching dateFrom and dateTo
	ngayFromStr := c.Query("dateFrom", "")
	ngayToStr := c.Query("dateTo", "")

	// --- Parse dates ---
	var ngayFrom, ngayTo time.Time
	var errTime error
	if ngayFromStr != "" {
		ngayFrom, errTime = time.Parse("2006-01-02", ngayFromStr)
		if errTime != nil {
			return service.SendError(c, 400, "Invalid format for ngayFrom (expect yyyy-mm-dd)")
		}
	}
	if ngayToStr != "" {
		ngayTo, errTime = time.Parse("2006-01-02", ngayToStr)
		if errTime != nil {
			return service.SendError(c, 400, "Invalid format for ngayTo (expect yyyy-mm-dd)")
		}
	}

	//calculating offset
	offset := (page-1)*6;

	//creating query mod
	queryMods := []qm.QueryMod{
		qm.Load(models.InvoiceRels.InvoiceStatusIDInvoiceStatus),
		qm.OrderBy("\"invoiceID\" DESC"),
	}

	// ==> Before adding qm.Limit and qm.Offset, create queryModsTotal
	queryModsTotal := []qm.QueryMod{}

	// handling search and sort filter and filling queryMods
	if search != "" {
		switch sort {
		case "Mã hóa đơn":
			queryMods = append(queryMods, qm.Where("CAST(\"invoiceID\" AS TEXT) ILIKE ?", "%"+search+"%"))
			queryModsTotal = append(queryModsTotal, qm.Where("CAST(\"invoiceID\" AS TEXT) ILIKE ?", "%"+search+"%"))
		case "Tên khách hàng":
			queryMods = append(queryMods, qm.Where("\"receiveName\" ILIKE ?", "%"+search+"%"))
			queryModsTotal = append(queryModsTotal, qm.Where("\"receiveName\" ILIKE ?", "%"+search+"%"))
		case "Trạng thái":
			status, err := models.InvoiceStatuses(qm.Where("\"statusName\" ILIKE ?", "%"+search+"%")).One(c.Context(), boil.GetContextDB())
			if err != nil {
				return service.SendError(c, 500, err.Error())
			}
			queryMods = append(queryMods, qm.Where("\"invoiceStatusID\" = ?", status.InvoiceStatusID))
			queryModsTotal = append(queryModsTotal, qm.Where("\"invoiceStatusID\" = ?", status.InvoiceStatusID))
		default:
			// fallback
		}
	} else if sort == "Ngày thanh toán" && !ngayFrom.IsZero() && !ngayTo.IsZero() && ngayFrom.Before(ngayTo) {
		queryMods = append(queryMods, qm.Where("DATE(\"paymentDate\") BETWEEN ? AND ?", ngayFrom, ngayTo))
		queryModsTotal = append(queryModsTotal, qm.Where("DATE(\"paymentDate\") BETWEEN ? AND ?", ngayFrom, ngayTo))
	}

	// Count total invoices matching filter
	totalInvoice, err := models.Invoices(queryModsTotal...).Count(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	// Tính lại totalPage
	totalPage := int(math.Ceil(float64(totalInvoice) / float64(6)))

	queryMods = append(queryMods, qm.Limit(6))
	queryMods = append(queryMods,qm.Offset(offset))

	invoices,err := models.Invoices(
		queryMods...
	).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	// Converting data into custom struct that has invoice, invoice details and its statuses
	response := make([]InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		response[i] = InvoiceResponse{
			Invoice: *invoice,
			InvoiceStatus: invoice.R.InvoiceStatusIDInvoiceStatus,
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

func GetAdminInvoiceDetail(c *fiber.Ctx) error{

	invoiceID := c.QueryInt("invoiceID",0);
	if invoiceID == 0{
		return service.SendError(c,400,"Did not receive invoiceID");
	}

	//get invoice first
	fetchInvoice, err := models.Invoices(
		qm.Where("\"invoiceID\" = ?",invoiceID),
	).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error())
	}
	//then get its status
	fetchStatus, err := models.InvoiceStatuses(qm.Where("\"invoiceStatusID\" = ?",fetchInvoice.InvoiceStatusID)).One(c.Context(),boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	queryMods := []qm.QueryMod{}

	switch(fetchStatus.StatusName){
		case "Đã đặt hàng":
			queryMods = append(queryMods, qm.Where("\"statusName\" LIKE ? OR \"statusName\" LIKE ? OR \"statusName\" LIKE ?",fetchStatus.StatusName,"Đã Hủy","Đã xác nhận"))
		case "Đã xác nhận":
			queryMods = append(queryMods, qm.Where("\"statusName\" LIKE ? OR \"statusName\" LIKE ?",fetchStatus.StatusName,"Đang xử lý"))
		case "Đang xử lý":
			queryMods = append(queryMods, qm.Where("\"statusName\" LIKE ? OR \"statusName\" LIKE ?",fetchStatus.StatusName,"Đang vận chuyển"))
		case "Đang vận chuyển":
			queryMods = append(queryMods, qm.Where("\"statusName\" LIKE ? OR \"statusName\" LIKE ?",fetchStatus.StatusName,"Giao thành công"))
		case "Giao thành công":
			queryMods = append(queryMods, qm.Where("\"statusName\" LIKE ?",fetchStatus.StatusName))
		default:
			//do nothing in fallback
	}

	listStatus,err := models.InvoiceStatuses(queryMods...).All(c.Context(),boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	//finally getting its details
	var listInvoiceDetails []InvoiceDetailResponse
	err2 := queries.Raw(`
		SELECT invoice_detail.*,
		product."productName" as food,
		invoice."receiveName" AS name, 
		invoice."receivePhone" AS phone, 
		invoice."receiveAddress" AS address
		FROM invoice 
		INNER JOIN invoice_detail ON invoice."invoiceID" = invoice_detail."invoiceID"
		INNER JOIN product ON invoice_detail."productID" = product."productID"
		WHERE invoice_detail."invoiceID" = $1
	`,fetchInvoice.InvoiceID).Bind(c.Context(),boil.GetContextDB(),&listInvoiceDetails);
	if err2 != nil{
		return service.SendError(c,500,err2.Error())
	}
	
	resp := fiber.Map{
		"status": "Success",
		"listStatus": listStatus,
		"listInvoiceDetails": listInvoiceDetails,
		"message": "Successfully fetched invoice detail values",
	}

	return c.JSON(resp);
}

type UpdateInvoiceStruct struct{
	StatusName string `json:"statusName"`
	CancelReason null.String `json:"cancelReason"`
}

func UpdateInvoice(c *fiber.Ctx) error{
	var status UpdateInvoiceStruct

	invoiceID := c.QueryInt("invoiceID",0);
	if invoiceID == 0{
		return service.SendError(c,400,"Did not receive invoiceID");
	}
	if err := c.BodyParser(&status); err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	invoiceStatus, err := models.InvoiceStatuses(qm.Where("\"statusName\" LIKE ?",status.StatusName)).One(c.Context(),boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	getInvoice,err := models.Invoices(
		qm.Where("\"invoiceID\" = ?",invoiceID),
	).One(c.Context(),boil.GetContextDB())
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	if getInvoice.InvoiceStatusID != invoiceStatus.InvoiceStatusID && invoiceStatus.InvoiceStatusID != 6{
		getInvoice.InvoiceStatusID += 1
	}

	if invoiceStatus.InvoiceStatusID == 6{
		getInvoice.InvoiceStatusID = 6
		getInvoice.CancelReason = status.CancelReason
	}
		
	_,err = getInvoice.Update(c.Context(),boil.GetContextDB(),boil.Infer())
	if err != nil{
		return service.SendError(c,500,err.Error())
	}

	resp := fiber.Map{
		"status": "Success",
		"invoice": getInvoice,
		"message": "Successfully updated invoice status!",
	}

	return c.JSON(resp);
}