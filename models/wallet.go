package models

import (
	"encoding/json"
)

// Wallet represents a currency wallet
type Wallet struct {
	WalletID string  `json:"wallet_id"`
	Balance  float64 `json:"balance"`
}

// Key returns the database key for this wallet
func (w *Wallet) Key() []byte {
	return []byte("wallet:" + w.WalletID)
}

// ToJSON converts the wallet to JSON
func (w *Wallet) ToJSON() ([]byte, error) {
	return json.Marshal(w)
}

// FromJSON populates the wallet from JSON
func (w *Wallet) FromJSON(data []byte) error {
	return json.Unmarshal(data, w)
}
