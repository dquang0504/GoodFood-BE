package handlers

import (
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
	if len(res.Candidates) > 0 && res.Candidates[0].Content != nil {
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
	order, err := models.Invoices(qm.Where("\"invoiceID\" = ?",orderID)).One(context.Background(),boil.GetContextDB())
	if err != nil{
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
