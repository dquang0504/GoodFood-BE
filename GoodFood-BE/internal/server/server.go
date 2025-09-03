package server

import (
	"github.com/gofiber/fiber/v2"

	"GoodFood-BE/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "GoodFood-BE",
			AppName:      "GoodFood-BE",
		}),

		db: database.New(),
	}

	return server
}
