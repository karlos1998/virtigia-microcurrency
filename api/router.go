package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"virtigia-microcurrency/db"
	"virtigia-microcurrency/middleware"
)

// SetupRouter sets up the router
func SetupRouter(database *db.DB) *gin.Engine {
	router := gin.Default()

	// Serve Swagger UI at root path
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create handler
	handler := NewHandler(database)

	// API routes
	api := router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		// Wallet routes
		wallets := api.Group("/wallets")
		{
			// Wallet operations
			wallets.POST("/:wallet_id/add", handler.AddCurrency)
			wallets.POST("/:wallet_id/remove", handler.RemoveCurrency)
			wallets.GET("/:wallet_id/balance", handler.GetWalletBalance)
			
			// Transaction history
			wallets.GET("/:wallet_id/transactions", handler.GetTransactionHistory)
		}
	}

	return router
}