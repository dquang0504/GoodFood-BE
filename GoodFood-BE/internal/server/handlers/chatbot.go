package handlers

import (
	"GoodFood-BE/internal/service"
	"GoodFood-BE/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"google.golang.org/genai"
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

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  "322745191572",
		Location: "us-central1",
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return service.SendError(c, 500, "Failed to create client: "+err.Error())
	}

	// ✅ Function declaration
	functions := []*genai.FunctionDeclaration{
		{
			Name: "get_order_status",
			Description: "Retrieve the status of an order in GoodFood24h",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"order_id": {Type: genai.TypeString, Description: "The ID of the order"},
				},
				Required: []string{"order_id"},
			},
		},
		{
			Name: "get_top_product",
			Description: "Retrieve the information of the most sold product in GoodFood24h",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
			},
		},
	}

	// ✅ Generation config with function calling
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: 1024,
		Temperature:     float32Ptr(0.7),
		TopP:            float32Ptr(0.9),
		Tools: []*genai.Tool{
			{FunctionDeclarations: functions},
		},
	}

	// ✅ Prompt
	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: "You are a concise assistant specialized in answering questions about GoodFood24h - an e-commerce website for ordering food online. " +
					"Answer this question or call a function if needed: " + body.Prompt},
			},
		},
	}

	// ✅ Generate content
	res, err := client.Models.GenerateContent(ctx,
		"projects/322745191572/locations/us-central1/endpoints/5530821664155107328",
		contents,
		config,
	)
	if err != nil {
		return service.SendError(c, 500, "Failed to generate content: "+err.Error())
	}

	// ✅ Debug full response
	debug, _ := json.MarshalIndent(res, "", "  ")
	log.Printf("Full response:\n%s\n", debug)

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
				// return c.JSON(fiber.Map{
				// 	"status":  "FunctionCall",
				// 	"message": "Model decided to call a function",
				// 	"function": fiber.Map{
				// 		"name":   call.Name,
				// 		"args":   call.Args,
				// 	},
				// })
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

func float32Ptr(f float32) *float32 {
	return &f
}

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
}

func get_top_product (c *fiber.Ctx) error{
	var response BestSellingProductResponse
	err := queries.Raw(`
		SELECT p."productID",
		p."productName",
		SUM(id."quantity") AS total_quantity_sold FROM
		invoice_detail id INNER JOIN product p ON
		id."productID" = p."productID"
		GROUP BY p."productID",p."productName"
		ORDER BY total_quantity_sold DESC
		LIMIT 1
	`).Bind(c.Context(),boil.GetContextDB(),&response);
	if err != nil{
		return service.SendError(c,500,err.Error());
	}
	
	result := fmt.Sprintf("The best selling product is: %s",response.ProductName)

	return c.JSON(fiber.Map{
		"status":  "Success",
		"data":    result,
		"message": "Fine-tuned model response OK",
	})
}
