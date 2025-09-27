package handlers

import (
	"GoodFood-BE/internal/auth"
	"GoodFood-BE/internal/dto"
	"GoodFood-BE/internal/service"
	"GoodFood-BE/internal/utils"
	"GoodFood-BE/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/gofiber/fiber/v2"
)

// CallVertexAI uses Vertex AI function calling to analyze prompt and execute one of the below functions
func CallVertexAI(c *fiber.Ctx) error {
	var body dto.PromptRequest
	if err := c.BodyParser(&body); err != nil {
		return service.SendError(c, 400, "Invalid prompt!")
	}

	if body.Prompt == "" {
		return service.SendError(c, 400, "Prompt cannot be empty")
	}

	res, err := utils.CallVertexAI(body.Prompt, c, true)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	fmt.Println(res)

	// If the model called a function
	if res != nil && len(res.Candidates) > 0 && res.Candidates[0] != nil {
		content := res.Candidates[0].Content
		if content != nil && len(content.Parts) > 0 {
			for _, part := range content.Parts {
				if part.FunctionCall != nil {
					call := part.FunctionCall
					if call.Args == nil {
						return service.SendError(c, 400, "Function call args is nil")
					}
					switch call.Name {
					case "get_order_status":
						orderIDStr := fmt.Sprintf("%v", call.Args["order_id"])
						orderID, err := strconv.Atoi(orderIDStr)
						if err != nil {
							return service.SendError(c, 400, "Invalid order_id in function call")
						}
						return get_order_status(c, orderID)
					case "get_top_product":
						return get_top_product(c)
					case "place_order":
						productsAny, ok := call.Args["products"]
						if !ok {
							return service.SendError(c, 400, "Missing products in function call")
						}
						products, ok := productsAny.([]interface{})
						if !ok {
							return service.SendError(c, 400, "Invalid products format")
						}
						var parsedProducts []map[string]interface{}
						for _, p := range products {
							m, ok := p.(map[string]interface{})
							if !ok {
								return service.SendError(c, 400, "Invalid product item format")
							}
							parsedProducts = append(parsedProducts, m)
						}
						paymentMethod, ok := call.Args["payment_method"].(string)
						if !ok {
							return service.SendError(c, 400, "Invalid payment_method")
						}
						return place_order(c, parsedProducts, paymentMethod)
					default:
						fmt.Println("Unknown function:", call.Name)
					}
				}
			}
		}
	}

	// Otherwise, return normal text
	result := ""
	if res != nil {
		for _, candidate := range res.Candidates {
			if candidate == nil || candidate.Content == nil {
				continue
			}
			if len(candidate.Content.Parts) == 0 {
				continue
			}
			for _, part := range candidate.Content.Parts {
				if part != nil && part.Text != "" {
					result += part.Text
				}
			}
		}
	}

	if result == "" {
		// Trả message fallback để tránh panic và dễ debug
		return service.SendError(c, 500, "Vertex AI returned empty or invalid response")
	}

	return c.JSON(fiber.Map{
		"status":  "Success",
		"data":    result,
		"message": "Fine-tuned model response OK",
	})
}

// get_order_status fetches the designated invoice status and return it to the user.
func get_order_status(c *fiber.Ctx, orderID int) error {

	//Fetch the currently logged in user
	username := auth.GetAuthenticatedUser(c)
	if username == "" {
		return service.SendError(c, 401, "User not authenticated")
	}
	account, err := models.Accounts(qm.Where("username = ?", username)).One(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Fetch the order
	order, err := models.Invoices(qm.Where("\"invoiceID\" = ? AND \"accountID\" = ?", orderID, account.AccountID)).One(c.Context(), boil.GetContextDB())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			resp, err := utils.GiveAnswerForUnreachableData("user asked for the status of an order, but the order does not exist in the database or it's not THEIR order", c)
			if err != nil {
				return service.SendError(c, 500, err.Error())
			}
			return c.JSON(fiber.Map{
				"status":  "Success",
				"data":    resp,
				"message": "Fine-tuned model response OK",
			})
		}
		return service.SendError(c, 500, err.Error())
	}

	//If found the order, fetch the order status.
	orderStatus, err := models.InvoiceStatuses(qm.Where("\"invoiceStatusID\" = ?", order.InvoiceStatusID)).One(context.Background(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	result := fmt.Sprintf("The status of order %d is: %s", orderID, orderStatus.StatusName)

	return c.JSON(fiber.Map{
		"status":  "Success",
		"data":    result,
		"message": "Fine-tuned model response OK",
	})
}

// get_top_product returns to the user the product with the most sales.
func get_top_product(c *fiber.Ctx) error {
	var response dto.BestSellingProductResponse
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
	`).Bind(c.Context(), boil.GetContextDB(), &response)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}
	res, err := utils.GiveStructuredAnswer("get_top_product", string(jsonBytes), c)
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	return c.JSON(fiber.Map{
		"status":    "Success",
		"data":      res,
		"image":     response.ProductImage,
		"productID": response.ProductID,
		"message":   "Fine-tuned model response OK",
	})
}

// This function handles user order placement logic.
// It validates authentication, retrieves user/address info,
// fetches product details concurrently, and builds the cart.
// Returns a JSON response with order summary or error message.
func place_order(c *fiber.Ctx, products []map[string]interface{}, paymentMethod string) error {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		carts []dto.CartDetailResponse
	)

	//Validate authentication
	username := auth.GetAuthenticatedUser(c)
	if username == "" {
		return service.SendError(c, 401, "User not authenticated")
	}
	user, err := models.Accounts(qm.Where("username = ?", username)).One(c.Context(), boil.GetContextDB())
	if err != nil {
		return service.SendError(c, 500, err.Error())
	}

	//Fetch delivery address
	address, err := models.Addresses(qm.Where("\"accountID\" = ? AND status = true", user.AccountID)).One(c.Context(), boil.GetContextDB())
	if err == sql.ErrNoRows {
		question := "The user wanted to place an order but they didn't have a delivery address. Politely asked them to add their delivery address first at Account > Delivery Address"
		resp, err := utils.GiveAnswerForUnreachableData(question, c)
		if err != nil {
			return service.SendError(c, 500, err.Error())
		}
		return c.JSON(fiber.Map{
			"status":  "Success",
			"data":    resp,
			"message": "Fine-tuned model response OK",
		})
	}

	//Channel to fetch err/resp from goroutines
	errChan := make(chan error, len(products))
	respChan := make(chan fiber.Map, 1)

	//Process products concurrently
	for _, pro := range products {
		wg.Add(1)
		go func(p map[string]interface{}) {
			defer wg.Done()

			quantity := int(p["quantity"].(float64))
			productName := p["product_name"].(string)
			product, err := models.Products(qm.Where("\"productName\" ILIKE ?", "%"+productName+"%")).One(c.Context(), boil.GetContextDB())
			if err == sql.ErrNoRows {
				question := fmt.Sprintf("User wanted to place an order which consists of a product that doesn't exist in the database, specifically they asked for %s", productName)
				resp, innerErr := utils.GiveAnswerForUnreachableData(question, c)
				if innerErr != nil {
					errChan <- innerErr
					return
				}
				//Send AI response to main goroutine
				respChan <- fiber.Map{
					"status":  "Success",
					"data":    resp,
					"message": "Fine-tuned model response OK",
				}
				return
			} else if err != nil {
				errChan <- err
				return
			}

			//Append items to cart safely
			mu.Lock()
			carts = append(carts, dto.CartDetailResponse{
				CartDetail: models.CartDetail{
					Quantity:  quantity,
					ProductID: product.ProductID,
					AccountID: user.AccountID,
				},
				Product: product,
			})
			mu.Unlock()
		}(pro)
	}

	//Wait for all goroutines to finish
	wg.Wait()
	close(errChan)
	close(respChan)

	//priority: if has resp from AI -> return resp
	if resp, ok := <-respChan; ok {
		return c.JSON(resp)
	}

	//return err if there are any
	for err := range errChan {
		if err != nil {
			return service.SendError(c, 500, err.Error())
		}
	}

	return c.JSON(fiber.Map{
		"status":        "Success",
		"data":          "Your order is ready. Click here!",
		"carts":         carts,
		"address":       address,
		"paymentMethod": paymentMethod,
		"message":       "Fine-tuned model response OK",
	})
}
