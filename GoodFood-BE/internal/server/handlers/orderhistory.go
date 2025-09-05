package handlers

import (
	redisdatabase "GoodFood-BE/internal/redis-database"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
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
	tab := c.Query("tab","Order Placed");
	accountID := c.QueryInt("accountID",0);
	if accountID == 0{
		return service.SendError(c,400,"Did not receive accountID!")
	}
	page := c.QueryInt("page",0);
	if page == 0{
		return service.SendError(c,400,"Did not receive pageNum");
	}

	fmt.Println(page);

	offset := (page - 1) * 6;

	//creating redis key
	redisKey := fmt.Sprintf("orderhistory:tab=%s:page=%d",tab,page);
	//checking if redis key exists
	cachedOrderHistory, err := redisdatabase.Client.Get(redisdatabase.Ctx,redisKey).Result();
	if err == nil{
		return c.JSON(json.RawMessage(cachedOrderHistory))
	}

	invoiceList := []InvoiceList{}
	err = queries.Raw(`
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
	totalPage := int(math.Ceil(float64(totalRecords) / 6.0))

	resp := fiber.Map{
		"status": "Success",
		"data": invoiceList,
		"totalPage": totalPage,
		"message": "Successfully fetched invoice list!",
	}

	//saving redis cache for 15 mins
	jsonData, _ := json.Marshal(resp);
	redisdatabase.Client.Set(redisdatabase.Ctx,redisKey,jsonData,15*time.Minute)
	//add this key to the set tracking all cache keys of this product
	redisSetKey := fmt.Sprintf("orderhistory:tab:%s",tab);
	err = redisdatabase.Client.SAdd(redisdatabase.Ctx,redisSetKey,redisKey).Err()
	if err != nil{
		fmt.Println("Error adding redis key to set: ", err)
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

	// Tabs that need to be renewed
	tabs := []string{"Order Placed", "Cancelled"}

	for _, t := range tabs {
		redisSetKey := fmt.Sprintf("orderhistory:tab:%s", t)
		keys, err := redisdatabase.Client.SMembers(redisdatabase.Ctx, redisSetKey).Result()
		if err != nil {
			fmt.Printf("Error getting keys from set %s: %v\n", redisSetKey, err)
			continue
		}

		if len(keys) > 0 {
			if err := redisdatabase.Client.Del(redisdatabase.Ctx, keys...).Err(); err != nil {
				fmt.Printf("Error deleting keys for tab %s: %v\n", t, err)
			}
		}
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