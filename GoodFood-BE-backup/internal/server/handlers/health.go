package handlers

import (
	"GoodFood-BE/internal/database"

	"github.com/gofiber/fiber/v2"
)

func HealthHandler(dbService database.Service) fiber.Handler {
	return func(c *fiber.Ctx) error{
		return c.JSON(dbService.Health())
	}
}