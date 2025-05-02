package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"virtigia-microcurrency/db"
)

func setupTestEnvironment(t *testing.T) (*gin.Engine, *db.DB, func()) {
	// Set test mode
	gin.SetMode(gin.TestMode)

	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test-db-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize database
	database, err := db.NewDB(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Set up router
	router := SetupRouter(database)

	// Set API token for tests
	os.Setenv("API_TOKEN", "test-token")

	// Return cleanup function
	cleanup := func() {
		database.Close()
		os.RemoveAll(tempDir)
		os.Unsetenv("API_TOKEN")
	}

	return router, database, cleanup
}

func TestAddCurrency(t *testing.T) {
	router, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// Create request
	req := AddCurrencyRequest{
		Amount:      100.0,
		Description: "Test deposit",
	}
	reqBody, _ := json.Marshal(req)

	// Create request
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/v1/wallets/"+walletID+"/add", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer test-token")

	// Perform request
	router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp TransactionResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Check response data
	assert.Equal(t, walletID, resp.Transaction.WalletID)
	assert.Equal(t, req.Amount, resp.Transaction.Amount)
	assert.Equal(t, req.Description, resp.Transaction.Description)
	assert.Equal(t, walletID, resp.Wallet.WalletID)
	assert.Equal(t, req.Amount, resp.Wallet.Balance)
}

func TestRemoveCurrency(t *testing.T) {
	router, db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// First add currency
	_, err := db.AddCurrency(walletID, 100.0, "Initial deposit", nil)
	assert.NoError(t, err)

	// Create request to remove currency
	req := RemoveCurrencyRequest{
		Amount:      50.0,
		Description: "Test withdrawal",
	}
	reqBody, _ := json.Marshal(req)

	// Create request
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/v1/wallets/"+walletID+"/remove", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer test-token")

	// Perform request
	router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp TransactionResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Check response data
	assert.Equal(t, walletID, resp.Transaction.WalletID)
	assert.Equal(t, -req.Amount, resp.Transaction.Amount) // Negative amount for removal
	assert.Equal(t, req.Description, resp.Transaction.Description)
	assert.Equal(t, walletID, resp.Wallet.WalletID)
	assert.Equal(t, 50.0, resp.Wallet.Balance) // 100 - 50 = 50
}

func TestGetWalletBalance(t *testing.T) {
	router, db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// Add some currency to the wallet
	_, err := db.AddCurrency(walletID, 100.0, "Initial deposit", nil)
	assert.NoError(t, err)

	// Create request
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/balance", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")

	// Perform request
	router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp WalletBalanceResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Check response data
	assert.Equal(t, walletID, resp.WalletID)
	assert.Equal(t, 100.0, resp.Balance)
}

func TestGetTransactionHistory(t *testing.T) {
	router, db, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// Add some transactions
	for i := 0; i < 5; i++ {
		_, err := db.AddCurrency(walletID, 10.0, "Test transaction", nil)
		assert.NoError(t, err)
	}

	// Create request
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")

	// Perform request
	router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp TransactionHistoryResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Check response data
	assert.Equal(t, walletID, resp.Wallet.WalletID)
	assert.Equal(t, 50.0, resp.Wallet.Balance) // 5 * 10 = 50
	assert.Equal(t, 5, len(resp.Transactions))
	assert.Equal(t, 5, resp.Pagination.Count)
}

func TestAuthMiddleware(t *testing.T) {
	router, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// Create request without token
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions", nil)

	// Perform request
	router.ServeHTTP(w, httpReq)

	// Check response - should be unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Create request with invalid token
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions", nil)
	httpReq.Header.Set("Authorization", "Bearer invalid-token")

	// Perform request
	router.ServeHTTP(w, httpReq)

	// Check response - should be unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}