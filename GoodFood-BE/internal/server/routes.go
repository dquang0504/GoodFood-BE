package server

import (
	"GoodFood-BE/internal/auth"
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
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: true, // credentials require explicit origins
		MaxAge:           300,
	}))

	//Route nhóm
	s.App.Get("/", handlers.HelloWorldHandler)
	s.App.Get("/health", handlers.HealthHandler(dbService))

	s.App.Get("/websocket", websocket.New(s.websocketHandler))

	//Routes related to accounts
	userGroup := s.App.Group("/api/user")
	userGroup.Get("/login",handlers.HandleLogin)
	userGroup.Get("/refresh-token",handlers.RefreshToken)
	//Routes related to products
	productGroup := s.App.Group("/api/products")
	productGroup.Get("/getFeaturings",handlers.GetFour)
	productGroup.Get("/getTypes",handlers.GetTypes)
	productGroup.Get("/",handlers.GetProductsByPage)
	productGroup.Post("/classify-image",handlers.ClassifyImage)
	productGroup.Get("/detail",handlers.GetDetail)
	productGroup.Get("/similar",handlers.GetSimilar)
	//Routes related to cart
	cartGroup := s.App.Group("/api/cart",auth.AuthMiddleware)
	cartGroup.Get("/fetch",handlers.FetchCart)
	cartGroup.Get("",handlers.GetCartDetail)
	cartGroup.Post("/modify",handlers.Cart_ModifyQuantity)
	cartGroup.Delete("/delete",handlers.DeleteCartItem)
	cartGroup.Post("/add",handlers.AddToCart)
	cartGroup.Delete("/deleteAll",handlers.DeleteAllItems)
	//Routes related to address
	addressGroup := s.App.Group("api/address",auth.AuthMiddleware)
	addressGroup.Get("/fetch",handlers.FetchAddress)
	addressGroup.Post("/insert",handlers.AddressInsert)
	addressGroup.Get("/detail",handlers.AddressDetail)
	addressGroup.Put("/update",handlers.AddressUpdate)
	addressGroup.Delete("/delete",handlers.AddressDelete)
	addressGroup.Get("/fill",handlers.AddressFill)
	addressGroup.Put("/quickChange",handlers.AddressQuickChange)
	//Routes related to invoice
	invoiceGroup := s.App.Group("api/invoice",auth.AuthMiddleware)
	invoiceGroup.Post("/pay",handlers.InvoicePay)
	//Routes related to Admin Dashboard
	dashboardGroup := s.App.Group("api/admin",auth.AuthMiddleware)
	dashboardGroup.Get("/dashboard",handlers.GetDashboard)
	dashboardGroup.Get("/linechart",handlers.GetLineChart)
	dashboardGroup.Get("/piechart",handlers.GetPieChart)
	dashboardGroup.Get("/barchart",handlers.GetBarChart)
	//Routes related to Admin Invoice
	adminInvoiceGroup := s.App.Group("api/admin/order",auth.AuthMiddleware)
	adminInvoiceGroup.Get("",handlers.GetAdminInvoice)
	adminInvoiceGroup.Get("/detail",handlers.GetAdminInvoiceDetail)
	adminInvoiceGroup.Put("/update",handlers.UpdateInvoice)
	//Routes related to Admin User
	adminUserGroup := s.App.Group("api/admin/user",auth.AuthMiddleware)
	adminUserGroup.Get("",handlers.GetAdminUsers)
	adminUserGroup.Get("/detail",handlers.GetAdminUserDetail)
	adminUserGroup.Post("/create",handlers.AdminUserCreate)
	adminUserGroup.Put("/update",handlers.AdminUserUpdate)
	//Routes related Admin Product Type
	adminProductTypeGroup := s.App.Group("api/admin/product-type",auth.AuthMiddleware)
	adminProductTypeGroup.Get("",handlers.GetAdminProductTypes)
	adminProductTypeGroup.Get("/detail",handlers.GetAdminProductTypeDetail)
	adminProductTypeGroup.Post("/create",handlers.AdminProductTypeCreate)
	adminProductTypeGroup.Put("/update",handlers.AdminProductTypeUpdate)
	//Routes related Admin Product
	adminProductGroup := s.App.Group("api/admin/product",auth.AuthMiddleware)
	adminProductGroup.Get("",handlers.GetAdminProducts)
	adminProductGroup.Get("/detail",handlers.GetAdminProductDetail);
	adminProductGroup.Post("/create",handlers.AdminProductCreate)
	adminProductGroup.Put("/update",handlers.AdminProductUpdate)
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
