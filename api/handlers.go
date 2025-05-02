package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"virtigia-microcurrency/db"
)

// Handler contains the handlers for the API
type Handler struct {
	DB *db.DB
}

// NewHandler creates a new Handler
func NewHandler(db *db.DB) *Handler {
	return &Handler{DB: db}
}

// AddCurrency adds currency to a wallet
// @Summary Add currency to a wallet
// @Description Add currency to a wallet and record the transaction
// @Tags wallet
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
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

	// Add currency to wallet
	tx, err := h.DB.AddCurrency(walletID, req.Amount, req.Description, req.AdditionalData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add currency: " + err.Error()})
		return
	}

	// Get updated wallet
	wallet, err := h.DB.GetWallet(walletID)
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

	// Remove currency from wallet
	tx, err := h.DB.RemoveCurrency(walletID, req.Amount, req.Description, req.AdditionalData)
	if err != nil {
		if err == db.ErrInsufficientFunds {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Insufficient funds"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove currency: " + err.Error()})
		return
	}

	// Get updated wallet
	wallet, err := h.DB.GetWallet(walletID)
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

	// Get wallet balance
	balance, err := h.DB.GetWalletBalance(walletID)
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
// @Param wallet_id path string true "Wallet ID"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
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

	// Get transactions
	transactions, err := h.DB.GetTransactionsByWallet(walletID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get transactions: " + err.Error()})
		return
	}

	// Get wallet
	wallet, err := h.DB.GetWallet(walletID)
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