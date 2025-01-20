package main

import (
	"GoodFood-BE/internal/database"
	"GoodFood-BE/internal/server"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

func gracefulShutdown(fiberServer *server.FiberServer, dbService database.Service, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := fiberServer.ShutdownWithContext(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	//Close database connection
	if err := dbService.Close(); err != nil{
		log.Printf("Error closing database: %v",err)
	}else{
		log.Println("Database connection closed.")
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	//Initialize database connection
	dbService := database.New()
	defer dbService.Close()

	//Check database health
	health := dbService.Health()
	log.Printf("Database Health: %+v\n",health)
	if health["status"] != "up"{
		log.Fatal("Database is not healthy. Exiting...")
	}

	//Initialize Fiber server
	server := server.New()

	//Register routes (pass dbService if needed in routes)
	server.RegisterFiberRoutes(dbService)

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	//Start a fiber server in a goroutine
	go func() {
		port, _ := strconv.Atoi(os.Getenv("PORT"))
		err := server.Listen(fmt.Sprintf(":%d", port))
		if err != nil {
			panic(fmt.Sprintf("http server error: %s", err))
		}
	}()

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, dbService ,done)

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
