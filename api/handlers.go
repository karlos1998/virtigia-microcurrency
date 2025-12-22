package api

import (
	"net/http"
	"strconv"

	"virtigia-microcurrency/db"
	"virtigia-microcurrency/middleware"

	"github.com/gin-gonic/gin"
)

// Handler contains the handlers for the API
type Handler struct {
	DBManager *db.DBManager
}

// NewHandler creates a new Handler
func NewHandler(dbManager *db.DBManager) *Handler {
	return &Handler{DBManager: dbManager}
}

// getDB returns the database for the current environment
func (h *Handler) getDB(c *gin.Context) (*db.DB, error) {
	env := middleware.GetEnvironment(c)
	return h.DBManager.GetDB(env)
}

// AddCurrency adds currency to a wallet
// @Summary Add currency to a wallet
// @Description Add currency to a wallet and record the transaction
// @Tags wallet
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param X-ENV header string false "Environment (default: production)"
// @Param wallet_id path string true "Wallet ID"
// @Param request body AddCurrencyRequest true "Add currency request"
// @Success 200 {object} TransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wallets/{wallet_id}/add [post]
func (h *Handler) AddCurrency(c *gin.Context) {
	walletID := c.Param("wallet_id")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Wallet ID is required"})
		return
	}

	var req AddCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request: " + err.Error()})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Amount must be positive"})
		return
	}

	// Get database for current environment
	database, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get database: " + err.Error()})
		return
	}

	// Add currency to wallet
	tx, err := database.AddCurrency(walletID, req.Amount, req.Description, req.AdditionalData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add currency: " + err.Error()})
		return
	}

	// Get updated wallet
	wallet, err := database.GetWallet(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get wallet: " + err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, TransactionResponse{
		Transaction: tx,
		Wallet:      wallet,
	})
}

// RemoveCurrency removes currency from a wallet
// @Summary Remove currency from a wallet
// @Description Remove currency from a wallet and record the transaction
// @Tags wallet
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param X-ENV header string false "Environment (default: production)"
// @Param wallet_id path string true "Wallet ID"
// @Param request body RemoveCurrencyRequest true "Remove currency request"
// @Success 200 {object} TransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wallets/{wallet_id}/remove [post]
func (h *Handler) RemoveCurrency(c *gin.Context) {
	walletID := c.Param("wallet_id")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Wallet ID is required"})
		return
	}

	var req RemoveCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request: " + err.Error()})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Amount must be positive"})
		return
	}

	// Get database for current environment
	database, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get database: " + err.Error()})
		return
	}

	// Remove currency from wallet
	tx, err := database.RemoveCurrency(walletID, req.Amount, req.Description, req.AdditionalData)
	if err != nil {
		if err == db.ErrInsufficientFunds {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Insufficient funds"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove currency: " + err.Error()})
		return
	}

	// Get updated wallet
	wallet, err := database.GetWallet(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get wallet: " + err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, TransactionResponse{
		Transaction: tx,
		Wallet:      wallet,
	})
}

// GetWalletBalance gets the balance of a wallet
// @Summary Get wallet balance
// @Description Get the balance of a wallet
// @Tags wallet
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param X-ENV header string false "Environment (default: production)"
// @Param wallet_id path string true "Wallet ID"
// @Success 200 {object} WalletBalanceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wallets/{wallet_id}/balance [get]
func (h *Handler) GetWalletBalance(c *gin.Context) {
	walletID := c.Param("wallet_id")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Wallet ID is required"})
		return
	}

	// Get database for current environment
	database, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get database: " + err.Error()})
		return
	}

	// Get wallet balance
	balance, err := database.GetWalletBalance(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get wallet balance: " + err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, WalletBalanceResponse{
		WalletID: walletID,
		Balance:  balance,
	})
}

// GetTransactionHistory gets the transaction history for a wallet
// @Summary Get transaction history
// @Description Get the transaction history for a wallet with pagination
// @Tags transactions
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param X-ENV header string false "Environment (default: production)"
// @Param wallet_id path string true "Wallet ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Param sort_by query string false "Sort by" default("timestamp")
// @Param sort_order query string false "Sort order" default("DESC")
// @Success 200 {object} TransactionHistoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /wallets/{wallet_id}/transactions [get]
func (h *Handler) GetTransactionHistory(c *gin.Context) {
	walletID := c.Param("wallet_id")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Wallet ID is required"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Parse sorting parameters
	sortBy := c.DefaultQuery("sort_by", "timestamp")
	sortOrder := c.DefaultQuery("sort_order", "DESC")

	// Validate sort_by parameter
	if sortBy != "timestamp" && sortBy != "amount" {
		sortBy = "timestamp"
	}

	// Validate and normalize sort_order parameter
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	// Get database for current environment
	database, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get database: " + err.Error()})
		return
	}

	// Get transactions
	transactions, err := database.GetTransactionsByWallet(walletID, limit, offset, sortBy, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get transactions: " + err.Error()})
		return
	}

	// Get wallet
	wallet, err := database.GetWallet(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get wallet: " + err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, TransactionHistoryResponse{
		Transactions: transactions,
		Wallet:       wallet,
		Pagination: Pagination{
			Limit:  limit,
			Offset: offset,
			Count:  len(transactions),
		},
	})
}
