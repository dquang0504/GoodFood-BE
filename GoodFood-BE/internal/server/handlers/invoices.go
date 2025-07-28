package handlers

import (
	"GoodFood-BE/config"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

//can make a custom InvoiceDetail slice that includes models.Product so that I can parse it from
//the payload

type InvoicePayload struct{
	Invoice models.Invoice `json:"invoice"`
	InvoiceDetails []models.InvoiceDetail `json:"invoiceDetails"`
	Products []models.Product `json:"product"`
}

func InvoicePay(c *fiber.Ctx) error{
	var payload InvoicePayload
	if err := c.BodyParser(&payload); err != nil{
		return service.SendError(c,401,"Invalid body details: " + err.Error());
	}

	//insert invoice first
	if err := payload.Invoice.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
		return service.SendError(c,500,err.Error());
	}

	//insert invoice details and link corresponding products
	for _,detail := range payload.InvoiceDetails{
		detail.InvoiceID = payload.Invoice.InvoiceID
		if err := detail.Insert(c.Context(),boil.GetContextDB(),boil.Infer()); err != nil{
			return service.SendError(c, 500, err.Error())
		}
	}
	
	resp := fiber.Map{
		"status": "Success",
		"data": payload,
		"message": "Successfully created new invoice!",
	}

	return c.JSON(resp);

}

func InvoiceOnlinePay(c *fiber.Ctx) error{

	body := models.Invoice{}
	if err := c.BodyParser(&body); err != nil{
		return service.SendError(c,400,err.Error());
	}

	//amount * 100 (vnpay requirement)
	amount := int64(body.TotalPrice) * 100
	orderId := strconv.Itoa(body.InvoiceID + 1)

	//query params
	vnpParams := map[string]string{
		"vnp_Version": 	 config.VnpVersion,
		"vnp_Command":   config.VnpCommand,
		"vnp_TmnCode":   config.VnpTmnCode,
		"vnp_Amount":    strconv.FormatInt(amount, 10),
		"vnp_CurrCode":  "VND",
		"vnp_BankCode":  "NCB",
		"vnp_TxnRef":    orderId,
		"vnp_OrderInfo": fmt.Sprintf("Paying for invoice: %d", body.InvoiceID),
		"vnp_Locale":    "vn",
		"vnp_OrderType": "other",
		"vnp_ReturnUrl": "https://fivefood.shop/vnpay_return",
		"vnp_IpAddr":    "127.0.0.1",
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

	//Build query string
	var query url.Values = url.Values{}
	var rawData string
	for i,k := range keys{
		v :=vnpParams[k]
		if v != ""{
			query.Add(k,v)
			rawData += k + "=" + v
			if i < len(keys) - 1{
				rawData += "&"
			}
		}
	}

	// Hash HMAC SHA512
	secureHash := config.HmacSHA512(config.SecretKey, rawData)
	query.Add("vnp_SecureHash", secureHash)

	// Build final URL
	paymentUrl := fmt.Sprintf("%s?%s", config.VnpPayURL, query.Encode())

	fmt.Println("VNPay RawData:", rawData)
	fmt.Println("VNPay SecureHash:", secureHash)


	resp := fiber.Map{
		"status": "Success",
		"data": paymentUrl,
		"message": "Successfully redirected to VNPay gateway!",
	}

	return c.JSON(resp);
}