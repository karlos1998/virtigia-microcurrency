package api

import (
	"virtigia-microcurrency/models"
)

// AddCurrencyRequest is the request for adding currency to a wallet
type AddCurrencyRequest struct {
	Amount        float64                `json:"amount" binding:"required,gt=0"`
	Description   string                 `json:"description" binding:"required"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// RemoveCurrencyRequest is the request for removing currency from a wallet
type RemoveCurrencyRequest struct {
	Amount        float64                `json:"amount" binding:"required,gt=0"`
	Description   string                 `json:"description" binding:"required"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// TransactionResponse is the response for a transaction
type TransactionResponse struct {
	Transaction *models.Transaction `json:"transaction"`
	Wallet      *models.Wallet      `json:"wallet"`
}

// TransactionHistoryResponse is the response for transaction history
type TransactionHistoryResponse struct {
	Transactions []*models.Transaction `json:"transactions"`
	Wallet       *models.Wallet        `json:"wallet"`
	Pagination   Pagination            `json:"pagination"`
}

// WalletBalanceResponse is the response for wallet balance
type WalletBalanceResponse struct {
	WalletID string  `json:"wallet_id"`
	Balance  float64 `json:"balance"`
}

// Pagination contains pagination information
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

// ErrorResponse is the response for an error
type ErrorResponse struct {
	Error string `json:"error"`
}