package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"virtigia-microcurrency/db"
)

func setupTestEnvironment(t *testing.T) (*gin.Engine, *db.DBManager, func()) {
	// Set test mode
	gin.SetMode(gin.TestMode)

	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test-db-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize database manager
	dbManager := db.NewDBManager(tempDir)

	// Set up router
	router := SetupRouter(dbManager)

	// Set API token for tests
	os.Setenv("API_TOKEN", "test-token")

	// Return cleanup function
	cleanup := func() {
		dbManager.Close()
		os.RemoveAll(tempDir)
		os.Unsetenv("API_TOKEN")
	}

	return router, dbManager, cleanup
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
	httpReq.Header.Set("X-ENV", "test")

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
	router, dbManager, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// Get database instance
	db, err := dbManager.GetDB("test")
	assert.NoError(t, err)

	// First add currency
	_, err = db.AddCurrency(walletID, 100.0, "Initial deposit", nil)
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
	httpReq.Header.Set("X-ENV", "test")

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
	router, dbManager, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// Get database instance
	db, err := dbManager.GetDB("test")
	assert.NoError(t, err)

	// Add some currency to the wallet
	_, err = db.AddCurrency(walletID, 100.0, "Initial deposit", nil)
	assert.NoError(t, err)

	// Create request
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/balance", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")
	httpReq.Header.Set("X-ENV", "test")

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
	router, dbManager, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// Get database instance
	db, err := dbManager.GetDB("test")
	assert.NoError(t, err)

	// Add some transactions
	for i := 0; i < 5; i++ {
		_, err := db.AddCurrency(walletID, 10.0, "Test transaction", nil)
		assert.NoError(t, err)
	}

	// Create request
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")
	httpReq.Header.Set("X-ENV", "test")

	// Perform request
	router.ServeHTTP(w, httpReq)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp TransactionHistoryResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
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

func TestGetTransactionHistorySortingByTimestamp(t *testing.T) {
	router, dbManager, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// Get database instance
	db, err := dbManager.GetDB("test")
	assert.NoError(t, err)

	// Add transactions with different timestamps (simulate by adding them sequentially)
	_, err = db.AddCurrency(walletID, 10.0, "Transaction 1", nil)
	assert.NoError(t, err)

	// Small delay to ensure different timestamps
	time.Sleep(1 * time.Millisecond)
	_, err = db.AddCurrency(walletID, 20.0, "Transaction 2", nil)
	assert.NoError(t, err)

	time.Sleep(1 * time.Millisecond)
	_, err = db.AddCurrency(walletID, 30.0, "Transaction 3", nil)
	assert.NoError(t, err)

	// Test DESC sorting (default)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions?sort_by=timestamp&sort_order=DESC", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")
	httpReq.Header.Set("X-ENV", "test")

	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp TransactionHistoryResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Should be sorted by timestamp DESC (newest first)
	assert.Equal(t, 3, len(resp.Transactions))
	assert.Equal(t, 30.0, resp.Transactions[0].Amount) // Newest first
	assert.Equal(t, 20.0, resp.Transactions[1].Amount)
	assert.Equal(t, 10.0, resp.Transactions[2].Amount)

	// Test ASC sorting
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions?sort_by=timestamp&sort_order=ASC", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")
	httpReq.Header.Set("X-ENV", "test")

	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Should be sorted by timestamp ASC (oldest first)
	assert.Equal(t, 3, len(resp.Transactions))
	assert.Equal(t, 10.0, resp.Transactions[0].Amount) // Oldest first
	assert.Equal(t, 20.0, resp.Transactions[1].Amount)
	assert.Equal(t, 30.0, resp.Transactions[2].Amount)
}

func TestGetTransactionHistorySortingByAmount(t *testing.T) {
	router, dbManager, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// Get database instance
	db, err := dbManager.GetDB("test")
	assert.NoError(t, err)

	// Add transactions with different amounts
	_, err = db.AddCurrency(walletID, 30.0, "Large transaction", nil)
	assert.NoError(t, err)

	_, err = db.AddCurrency(walletID, 10.0, "Small transaction", nil)
	assert.NoError(t, err)

	_, err = db.AddCurrency(walletID, 20.0, "Medium transaction", nil)
	assert.NoError(t, err)

	// Test DESC sorting by amount
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions?sort_by=amount&sort_order=DESC", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")
	httpReq.Header.Set("X-ENV", "test")

	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp TransactionHistoryResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Should be sorted by amount DESC (highest first)
	assert.Equal(t, 3, len(resp.Transactions))
	assert.Equal(t, 30.0, resp.Transactions[0].Amount) // Highest first
	assert.Equal(t, 20.0, resp.Transactions[1].Amount)
	assert.Equal(t, 10.0, resp.Transactions[2].Amount)

	// Test ASC sorting by amount
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions?sort_by=amount&sort_order=ASC", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")
	httpReq.Header.Set("X-ENV", "test")

	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Should be sorted by amount ASC (lowest first)
	assert.Equal(t, 3, len(resp.Transactions))
	assert.Equal(t, 10.0, resp.Transactions[0].Amount) // Lowest first
	assert.Equal(t, 20.0, resp.Transactions[1].Amount)
	assert.Equal(t, 30.0, resp.Transactions[2].Amount)
}

func TestGetTransactionHistoryPaginationWithSorting(t *testing.T) {
	router, dbManager, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// Get database instance
	db, err := dbManager.GetDB("test")
	assert.NoError(t, err)

	// Add multiple transactions
	for i := 1; i <= 10; i++ {
		_, err := db.AddCurrency(walletID, float64(i*10), "Transaction "+strconv.Itoa(i), nil)
		assert.NoError(t, err)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// Test pagination with sorting DESC by amount, limit 3, offset 2
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions?sort_by=amount&sort_order=DESC&limit=3&offset=2", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")
	httpReq.Header.Set("X-ENV", "test")

	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp TransactionHistoryResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Should return 3 transactions, starting from offset 2 in sorted list
	assert.Equal(t, 3, len(resp.Transactions))
	assert.Equal(t, 3, resp.Pagination.Count)
	assert.Equal(t, 3, resp.Pagination.Limit)
	assert.Equal(t, 2, resp.Pagination.Offset)

	// Should be sorted by amount DESC and paginated correctly
	// Full sorted list would be: [100, 90, 80, 70, 60, 50, 40, 30, 20, 10]
	// With offset 2, limit 3: [80, 70, 60]
	assert.Equal(t, 80.0, resp.Transactions[0].Amount)
	assert.Equal(t, 70.0, resp.Transactions[1].Amount)
	assert.Equal(t, 60.0, resp.Transactions[2].Amount)
}

func TestGetTransactionHistoryEdgeCases(t *testing.T) {
	router, dbManager, cleanup := setupTestEnvironment(t)
	defer cleanup()

	walletID := "wallet123"

	// Get database instance
	db, err := dbManager.GetDB("test")
	assert.NoError(t, err)

	// Test with empty wallet (no transactions)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")
	httpReq.Header.Set("X-ENV", "test")

	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp TransactionHistoryResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(resp.Transactions))
	assert.Equal(t, 0, resp.Pagination.Count)

	// Add one transaction
	_, err = db.AddCurrency(walletID, 50.0, "Single transaction", nil)
	assert.NoError(t, err)

	// Test invalid sort_by parameter (should default to timestamp)
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions?sort_by=invalid", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")
	httpReq.Header.Set("X-ENV", "test")

	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(resp.Transactions))
	assert.Equal(t, 50.0, resp.Transactions[0].Amount)

	// Test invalid sort_order parameter (should default to DESC)
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions?sort_order=invalid", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")
	httpReq.Header.Set("X-ENV", "test")

	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(resp.Transactions))
	assert.Equal(t, 50.0, resp.Transactions[0].Amount)

	// Test offset beyond available data
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("GET", "/api/v1/wallets/"+walletID+"/transactions?offset=10", nil)
	httpReq.Header.Set("Authorization", "Bearer test-token")
	httpReq.Header.Set("X-ENV", "test")

	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(resp.Transactions))
	assert.Equal(t, 0, resp.Pagination.Count)
}
