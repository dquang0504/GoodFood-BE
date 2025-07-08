package handlers

import (
	"GoodFood-BE/internal/auth"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type PromptRequest struct {
	Prompt string `json:"prompt"`
}

func CallVertexAI(c *fiber.Ctx) error {
	var body PromptRequest
	if err := c.BodyParser(&body); err != nil {
		return service.SendError(c, 400, "Invalid prompt!")
	}

	if body.Prompt == "" {
		return service.SendError(c, 400, "Prompt cannot be empty")
	}

	res,err := service.CallVertexAI(body.Prompt,c,true);
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	// ✅ If the model called a function
	if len(res.Candidates) > 0 && res.Candidates[0].Content != nil{
		for _, part := range res.Candidates[0].Content.Parts {
			if part.FunctionCall != nil {
				// The model decided to call a function
				call := part.FunctionCall
				switch(call.Name){
					case "get_order_status":
						orderIDStr := fmt.Sprintf("%v", call.Args["order_id"])
						orderID, err := strconv.Atoi(orderIDStr)
						if err != nil {
							return service.SendError(c, 400, "Invalid order_id in function call")
						}
						return get_order_status(c,orderID)
					case "get_top_product":
						return get_top_product(c);
					case "place_order":
						// Parse products
						productsAny := call.Args["products"]
						products, ok := productsAny.([]interface{})
						if !ok {
							return service.SendError(c, 400, "Invalid products format.")
						}
						var parsedProducts []map[string]interface{}
						for _, p := range products {
							m, ok := p.(map[string]interface{})
							if !ok {
								return service.SendError(c, 400, "Invalid product item format.")
							}
							parsedProducts = append(parsedProducts, m)
						}
						//Parse address_id
						var addressID int
						//checking if the model did return the address_id
						addrAny, ok := call.Args["address_id"];
						if ok{
							addrFloat, ok := addrAny.(float64)
							if !ok {
								return service.SendError(c, 400, "Invalid address_id.")
							}
							addressID = int(addrFloat)
						}else{
							addressID = 0;
						}
						
						// Parse payment_method
						paymentMethod, ok := call.Args["payment_method"].(string)
						if !ok {
							return service.SendError(c, 400, "Invalid payment_method.")
						}
						return place_order(c,parsedProducts,addressID,paymentMethod);
					default:
						break;
				}
			}
		}
	}

	// ✅ Otherwise, return normal text
	result := ""
	for _, candidate := range res.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				result += part.Text
			}
		}
	}

	return c.JSON(fiber.Map{
		"status":  "Success",
		"data":    result,
		"message": "Fine-tuned model response OK",
	})
}

//fix cho get_order_status chỉ trả về những hóa đơn của ng đang đăng nhập và suy nghĩ thêm các function mới
func get_order_status (c *fiber.Ctx, orderID int) error{
	account, err := models.Accounts(qm.Where("username = ?",auth.GetAuthenticatedUser(c))).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	
	order, err := models.Invoices(qm.Where("\"invoiceID\" = ? AND \"accountID\" = ?",orderID,account.AccountID)).One(c.Context(),boil.GetContextDB())
	if err != nil{
		if err.Error() == "sql: no rows in result set"{
			resp, err := service.GiveAnswerForUnreachableData("user asked for the status of an order, but the order does not exist in the database or it's not THEIR order",c)
			if err != nil{
				return service.SendError(c,500,err.Error());
			}
			return c.JSON(fiber.Map{
				"status":  "Success",
				"data":    resp,
				"message": "Fine-tuned model response OK",
			})
		}
		return service.SendError(c,500,err.Error());
	}
	
	orderStatus, err := models.InvoiceStatuses(qm.Where("\"invoiceStatusID\" = ?",order.InvoiceStatusID)).One(context.Background(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	
	result := fmt.Sprintf("The status of order %d is: %s",orderID,orderStatus.StatusName)

	return c.JSON(fiber.Map{
		"status":  "Success",
		"data":    result,
		"message": "Fine-tuned model response OK",
	})
}

type BestSellingProductResponse struct {
	ProductID         int    `boil:"productID" json:"productID"`
	ProductName       string `boil:"productName" json:"productName"`
	TotalQuantitySold int    `boil:"total_quantity_sold" json:"totalQuantitySold"`
	ProductImage string `boil:"product_img" json:"productImage"`
}

func get_top_product (c *fiber.Ctx) error{
	var response BestSellingProductResponse
	err := queries.Raw(`
		SELECT p."productID",
		p."productName",
		SUM(id."quantity") AS total_quantity_sold,
		p."coverImage" AS product_img FROM
		invoice_detail id INNER JOIN product p ON
		id."productID" = p."productID"
		GROUP BY p."productID",p."productName",p."coverImage"
		ORDER BY total_quantity_sold DESC
		LIMIT 1
	`).Bind(c.Context(),boil.GetContextDB(),&response);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	jsonBytes, err := json.Marshal(response);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	res, err := service.GiveStructuredAnswer("get_top_product",string(jsonBytes),c);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	return c.JSON(fiber.Map{
		"status":  "Success",
		"data":    res,
		"image": response.ProductImage,
		"productID": response.ProductID,
		"message": "Fine-tuned model response OK",
	})
}

func place_order(c *fiber.Ctx, products []map[string]interface{}, addressID int, paymentMethod string) error{
	var address *models.Address
	carts := []models.CartDetail{}
	user, err := models.Accounts(qm.Where("username = ?",auth.GetAuthenticatedUser(c))).One(c.Context(),boil.GetContextDB());
	if err != nil{
		return service.SendError(c,500,err.Error());
	}

	//address fetching logic
	if addressID == 0{
		address, err = models.Addresses(qm.Where("\"accountID\" = ? AND status = true",user.AccountID)).One(c.Context(),boil.GetContextDB());
		if err != nil{
			return service.SendError(c,500,err.Error());
		}
	}else{
		address, err = models.Addresses(qm.Where("\"addressID\" = ? AND \"accountID\" = ?",addressID,user.AccountID)).One(c.Context(),boil.GetContextDB());
		if err != nil{
			return service.SendError(c,500,err.Error());
		}
	}

	//products fetching logic
	for _,pro := range products{
		quantity := int(pro["quantity"].(float64))
		productName := pro["product_name"].(string)
		product, err := models.Products(qm.Where("\"productName\" ILIKE ?","%"+productName+"%")).One(c.Context(),boil.GetContextDB());
		if err != nil{
			return service.SendError(c,500,err.Error());
		}

		//inserting product to cart
		carts = append(carts, models.CartDetail{Quantity: quantity,ProductID: product.ProductID,AccountID: user.AccountID})
	}

	return c.JSON(fiber.Map{
		"status":  "Success",
		"data":    "Your order is ready. Click here!",
		"carts": carts,
		"address": address,
		"message": "Fine-tuned model response OK",
	})
}
