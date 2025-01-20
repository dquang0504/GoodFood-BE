package server

import (
	"GoodFood-BE/internal/database"
	"GoodFood-BE/internal/server/handlers"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/gofiber/contrib/websocket"
)

func (s *FiberServer) RegisterFiberRoutes(dbService database.Service) {
	// Apply CORS middleware
	s.App.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: false, // credentials require explicit origins
		MaxAge:           300,
	}))

	//Route nhóm
	s.App.Get("/", handlers.HelloWorldHandler)
	s.App.Get("/health", handlers.HealthHandler(dbService))

	s.App.Get("/websocket", websocket.New(s.websocketHandler))

	//Nhóm route liên quan đến user
	userGroup := s.App.Group("/api/user")
	userGroup.Get("/login",handlers.HandleLogin)
	//Nhóm route liên quan đến product
	productGroup := s.App.Group("/api/products")
	productGroup.Get("/getFeaturings",handlers.GetFour)
	productGroup.Get("/getTypes",handlers.GetTypes)
	productGroup.Get("/",handlers.GetProductsByPage)

}

func (s *FiberServer) websocketHandler(con *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			_, _, err := con.ReadMessage()
			if err != nil {
				cancel()
				log.Println("Receiver Closing", err)
				break
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			payload := fmt.Sprintf("server timestamp: %d", time.Now().UnixNano())
			if err := con.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
				log.Printf("could not write to socket: %v", err)
				return
			}
			time.Sleep(time.Second * 2)
		}
	}
}
