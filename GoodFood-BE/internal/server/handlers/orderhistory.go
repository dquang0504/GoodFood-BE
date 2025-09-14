package handlers

import (
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/gofiber/fiber/v2"
)

//GetOrderHistory returns a paginated order history list and caches the data into redis for later usages.
func GetOrderHistory(c *fiber.Ctx) error{
	//Fetch query params
	tab := c.Query("tab","Order Placed");
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive accountID!")
	}
	page := c.QueryInt("page",0);
	if page == 0{
		return service.SendError(c,400,"Did not receive pageNum");
	}

	//Create redis key
	redisKey := fmt.Sprintf("orderhistory:tab=%s:page=%d",tab,page);
	//Fetch cache
	cachedOrderHistory := fiber.Map{}
	if ok, _ := utils.GetCache(redisKey,&cachedOrderHistory); ok{
		return c.JSON(cachedOrderHistory)
	}

	//Have to calculate offset before fetching invoiceList
	offset := (page - 1) * utils.PageSize;
	invoiceList := []dto.InvoiceList{}
	err := queries.Raw(`
		SELECT invoice_detail."invoiceID" as invoice_id, COALESCE(COUNT(invoice_detail."productID"),0) as total_products,
		invoice."receiveAddress" as address, invoice.status as status, invoice."totalPrice" as total_money,
		invoice."cancelReason" as cancel_reason
		FROM invoice INNER JOIN invoice_detail
		ON invoice."invoiceID" = invoice_detail."invoiceID"
		INNER JOIN invoice_status ON invoice."invoiceStatusID" = invoice_status."invoiceStatusID"
		WHERE invoice_status."statusName" = $1 AND invoice."accountID" = $2
		GROUP BY invoice_detail."invoiceID", invoice."receiveAddress", invoice.status, invoice."totalPrice", invoice."cancelReason"
		LIMIT 6
		OFFSET $3
	`,tab,accountID,offset).Bind(c.Context(),boil.GetContextDB(),&invoiceList);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	//calculating totalPage
	var totalRecords int
	err = queries.Raw(`
		SELECT COUNT(DISTINCT invoice_detail."invoiceID")
		FROM invoice
		INNER JOIN invoice_detail
			ON invoice."invoiceID" = invoice_detail."invoiceID"
		INNER JOIN invoice_status
			ON invoice."invoiceStatusID" = invoice_status."invoiceStatusID"
		WHERE invoice_status."statusName" = $1
		AND invoice."accountID" = $2
	`, tab, accountID).QueryRow(boil.GetContextDB()).Scan(&totalRecords)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}
	_, totalPage := utils.Paginate(page,utils.PageSize,totalRecords);

	resp := fiber.Map{
		"status": "Success",
		"data": invoiceList,
		"totalPage": totalPage,
		"message": "Successfully fetched invoice list!",
	}

	//saving redis cache for 15 mins
	redisSetKey := fmt.Sprintf("orderhistory:tab:%s",tab);
	utils.SetCache(redisKey,resp,15*time.Minute,redisSetKey);

	return c.JSON(resp);
}

//CancelOrder changes the status of an order from "Order Placed" to "Cancelled" along with cancelReason.
func CancelOrder(c *fiber.Ctx) error{
	//Fetch query param
	invoiceID := c.QueryInt("invoiceID",0);
	if invoiceID == 0{
		return service.SendError(c,400, "Did not receive invoiceID!");
	}

	var invoice dto.InvoiceList
	if err := c.BodyParser(&invoice); err != nil{
		return service.SendError(c,400,err.Error());
	}

	//Update
	toUpdate, err := models.FindInvoice(c.Context(),boil.GetContextDB(),invoiceID)
	if err != nil {
		return service.SendError(c,500,err.Error());
	}
	toUpdate.InvoiceStatusID = 6
	toUpdate.CancelReason = null.StringFrom(invoice.CancelReason)
	if _ ,err = toUpdate.Update(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}

	//Caches that need to be renewed
	utils.ClearCache("orderhistory:tab:Order Placed","orderhistory:tab:Cancelled");

	resp := fiber.Map{
		"status": "Success",
		"data": toUpdate,
		"message": "Successfully canceled the order!",
	}

	return c.JSON(resp);
}

//GetOrderHistoryDetails returns the details of an invoice when clicked on.
func GetOrderHistoryDetail(c *fiber.Ctx) error{
	//Fetch query param
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

	//Response mapping
	response := make([]dto.InvoiceDetailStruct,len(invoiceDetails));
	for i, detail := range invoiceDetails{
		reviewExists, _ := models.Reviews(
			qm.Where("\"productID\" = ?",detail.R.ProductIDProduct.ProductID),
		).Exists(c.Context(),boil.GetContextDB());

		response[i] = dto.InvoiceDetailStruct{
			InvoiceID: detail.InvoiceID,
			Image: detail.R.ProductIDProduct.CoverImage,
			Product: *detail.R.ProductIDProduct,
			Quantity: detail.Quantity,
			TotalMoney: float64(detail.Price),
			ShippingFee: float64(detail.R.InvoiceIDInvoice.ShippingFee),
			ReviewCheck: reviewExists,
		}
	}

	resp := fiber.Map{
		"status": "Success",
		"data": response,
		"message": "Successfully fetched invoice details!",
	}

	return c.JSON(resp);
}