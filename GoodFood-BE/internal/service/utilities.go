package service

import "github.com/gofiber/fiber/v2"

func SendError(c *fiber.Ctx, statusCode int, message string) error{
	return c.Status(statusCode).JSON(fiber.Map{
		"status": "error",
		"message": message,
	})
}

func SendJSON(c *fiber.Ctx,status string, data interface{}, extras map[string]interface{},message string) error{
	
	//creating base response
	resp := fiber.Map{
		"status": status,
		"data": data,
		"message": message,
	}

	//adding extra variables
	for key, value := range extras{
		resp[key] = value
	}

	return c.JSON(resp)
}