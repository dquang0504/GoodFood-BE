package handlers

import (
	"GoodFood-BE/config"
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/gofiber/fiber/v2"
)

//InvoicePay creates an invoice for the purchased products and insert data into related tables
func InvoicePay(c *fiber.Ctx) error{
	var payload dto.InvoicePayload
	if err := c.BodyParser(&payload); err != nil{
		return service.SendError(c,401,"Invalid body details: " + err.Error());
	}

	//Open transaction
	tx, err := boil.BeginTx(c.Context(),nil);
	if err != nil{
		return service.SendError(c,500,err.Error())
	}
	defer tx.Rollback() //rollback data if something went wrong

	//Insert invoice first
	if err := payload.Invoice.Insert(c.Context(),tx,boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}

	//Insert invoice details
	for _,detail := range payload.InvoiceDetails{
		detail.InvoiceID = payload.Invoice.InvoiceID
		if err := detail.Insert(c.Context(),tx,boil.Infer()); err != nil{
			return service.SendError(c, 500, err.Error())
		}
	}
	
	//Commit transaction
	if err := tx.Commit(); err != nil{
		return service.SendError(c,500,err.Error())
	}
	
	resp := fiber.Map{
		"status": "Success",
		"data": payload,
		"message": "Successfully created new invoice!",
	}

	return c.JSON(resp);
}

//InvoicePayVNPAY receives invoice details from front-end, then construct a payment url to VNPAY
func InvoicePayVNPAY(c *fiber.Ctx) error{
	body := models.Invoice{}
	if err := c.BodyParser(&body); err != nil{
		return service.SendError(c,400,err.Error());
	}

	//amount * 100 (vnpay requirement)
	amount := int64(body.TotalPrice) * 100
	// Fetch latest invoiceID from db
	var latestID int
	err := boil.GetContextDB().QueryRowContext(c.Context(), `
		SELECT COALESCE(MAX("invoiceID"), 0) FROM invoice
	`).Scan(&latestID)
	if err != nil {
		return service.SendError(c, 500, "Failed to fetch latest invoiceID: "+err.Error())
	}
	orderId := strconv.Itoa(latestID + 1) //unique orderID for vnpay payment

	//query params
	vnpParams := map[string]string{
		"vnp_Version": 	 config.VnpVersion,
		"vnp_Command":   config.VnpCommand,
		"vnp_TmnCode":   os.Getenv("VNPAY_TMN"),
		"vnp_Amount":    strconv.FormatInt(amount, 10),
		"vnp_CurrCode":  "VND",
		"vnp_BankCode":  "NCB",
		"vnp_TxnRef":    orderId,
		"vnp_OrderInfo": fmt.Sprintf("Paying for invoice: %d", body.InvoiceID),
		"vnp_Locale":    "vn",
		"vnp_OrderType": "other",
		"vnp_ReturnUrl": os.Getenv("VNPAY_RETURN_URL"),
		"vnp_IpAddr":    c.IP(),
	}

	//Time zone in Asia/HCM
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	now := time.Now().In(loc)
	vnpParams["vnp_CreateDate"] = now.Format("20060102150405")
	vnpParams["vnp_ExpireDate"] = now.Add(15 * time.Minute).Format("20060102150405")

	// Sort keys before encoding
	var keys []string
	for k := range vnpParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	//Build rawData and query string
	var rawData strings.Builder
	var query strings.Builder
	for i, k := range keys {
		v := vnpParams[k]
		if v != "" {
			if i > 0 {
				rawData.WriteString("&")
				query.WriteString("&")
			}
			// phải encode value theo chuẩn VNPay
			encodedVal := url.QueryEscape(v)
			rawData.WriteString(k + "=" + encodedVal)
			query.WriteString(k + "=" + encodedVal)
		}
	}

	// Hash HMAC SHA512
	secureHash := config.HmacSHA512(os.Getenv("VNPAY_SECRET"), rawData.String())

	// Thêm SecureHash vào cuối query (KHÔNG encode lại chuỗi hash)
	query.WriteString("&vnp_SecureHash=" + secureHash)

	// Build final URL
	paymentUrl := fmt.Sprintf("%s?%s", config.VnpPayURL, query.String())

	resp := fiber.Map{
		"status": "Success",
		"data": paymentUrl,
		"message": "Successfully redirected to VNPay gateway!",
	}

	return c.JSON(resp);
}
