package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"math"
	"github.com/gofiber/fiber/v2"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

//GetAdminInvoice fetches invoices with filters, pagination, and summary metrics.
//Has a filter processing logic and returns the final list of data with pagination.
func GetAdminInvoice(c *fiber.Ctx) error{
	//Fetch metrics for InvoiceCards
	cards, err := utils.FetchInvoiceCards(c);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	//Fetch query params
	page := c.QueryInt("page",0);
	if page == 0{
		return service.SendError(c,401,"Did not receive page");
	}
	sort := c.Query("sort","");
	search := c.Query("search","")
	dateFrom, dateTo, err := utils.ParseDateRange(c.Query("dateFrom", ""), c.Query("dateTo", ""));
	if err != nil{
		return service.SendError(c,400,err.Error());
	}

	//Build query modifiers (both for fetching data and counting)
	queryMods, queryModsTotal, err := utils.BuildInvoiceFilters(c,search,sort,dateFrom,dateTo);
	if err != nil{
		return service.SendError(c,500, err.Error());
	}

	// Count total invoices that match queryModsTotal
	totalInvoice, err := models.Invoices(queryModsTotal...).Count(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}
	totalPage := int(math.Ceil(float64(totalInvoice) / float64(6)))

	//Add pagination
	offset := (page-1)*6;
	queryMods = append(queryMods, qm.Limit(6), qm.Offset(offset))

	//Fetch invoices with queryMods
	invoices,err := models.Invoices(queryMods...).All(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	// Converting to response format
	response := utils.MapInvoices(invoices);

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"cards": cards,
		"totalPage": totalPage,
		"message": "Successfully fetched invoice values",
	}

	return c.JSON(resp);
}

// GetAdminInvoiceDetail fetches invoice detail data including invoice status progression and detailed line items.
func GetAdminInvoiceDetail(c *fiber.Ctx) error{
	invoiceID := c.QueryInt("invoiceID",0);
	if invoiceID == 0{
		return service.SendError(c,400,"Did not receive invoiceID");
	}

	//Load invoice and current status
	invoice,status, err := utils.FetchInvoiceAndStatus(c,invoiceID);
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Determine possible status progression
	statusList, err := utils.FetchStatusProgression(c,status);
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Load invoice detail lines with joins
	details, err := utils.FetchInvoiceDetails(c, invoice.InvoiceID);
	if err != nil{
		return service.SendError(c,500,err.Error())
	}
	
	resp := fiber.Map{
		"status": "Success",
		"listStatus": statusList,
		"listInvoiceDetails": details,
		"message": "Successfully fetched invoice detail values",
	}

	return c.JSON(resp);
}

// UpdateInvoice updates an invoice's status and returns the updateđ invocie.
func UpdateInvoice(c *fiber.Ctx) error{
	var status dto.UpdateInvoiceStruct

	invoiceID := c.QueryInt("invoiceID",0);
	if invoiceID == 0{
		return service.SendError(c,400,"Did not receive invoiceID");
	}
	if err := c.BodyParser(&status); err != nil{
		return service.SendError(c,400,"Invalid body!");
	}

	getInvoice, err := utils.UpdateInvoiceStatus(c,invoiceID,status);
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