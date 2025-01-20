package service

import "github.com/gofiber/fiber/v2"

func SendError(c *fiber.Ctx, statusCode int, message string) error{
	return c.Status(statusCode).JSON(fiber.Map{
		"status": "error",
		"message": message,
	})
}