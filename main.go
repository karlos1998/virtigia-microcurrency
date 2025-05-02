package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"virtigia-microcurrency/api"
	"virtigia-microcurrency/db"
	_ "virtigia-microcurrency/docs"
)

// @title Virtigia Microcurrency API
// @version 1.0
// @description A lightweight microservice for handling in-game microcurrency transactions.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.virtigia.com/support
// @contact.email support@virtigia.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8880
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the API token.

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get data directory from environment or use default
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(".", "data")
	}

	// Initialize database manager
	dbManager := db.NewDBManager(dataDir)
	defer dbManager.Close()

	// Set up router
	router := api.SetupRouter(dbManager)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8880"
	}

	// Create server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
