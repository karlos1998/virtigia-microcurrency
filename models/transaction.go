package models

import (
	"encoding/json"
	"time"
)

// Transaction represents a currency transaction in the system
type Transaction struct {
	ID            string                 `json:"id"`
	WalletID      string                 `json:"wallet_id"`
	Amount        float64                `json:"amount"`
	Description   string                 `json:"description"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// Key returns the database key for this transaction
func (t *Transaction) Key() []byte {
	return []byte("transaction:" + t.ID)
}

// WalletKey returns the key for indexing by wallet ID
func (t *Transaction) WalletKey() []byte {
	return []byte("wallet:" + t.WalletID + ":transaction:" + t.ID)
}

// ToJSON converts the transaction to JSON
func (t *Transaction) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON populates the transaction from JSON
func (t *Transaction) FromJSON(data []byte) error {
	return json.Unmarshal(data, t)
}
