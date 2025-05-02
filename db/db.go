package db

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v3"
	"virtigia-microcurrency/models"
)

var (
	// ErrNotFound is returned when a record is not found
	ErrNotFound = errors.New("record not found")

	// ErrInsufficientFunds is returned when a wallet doesn't have enough balance
	ErrInsufficientFunds = errors.New("insufficient funds")
)

// DB represents the database
type DB struct {
	db *badger.DB
}

// NewDB creates a new database instance
func NewDB(dataDir string) (*DB, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	options := badger.DefaultOptions(dataDir)
	options.Logger = nil // Disable logging

	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

// Close closes the database
func (d *DB) Close() error {
	return d.db.Close()
}

// GetWallet retrieves a wallet by wallet ID
func (d *DB) GetWallet(walletID string) (*models.Wallet, error) {
	wallet := &models.Wallet{WalletID: walletID}

	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(wallet.Key())
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return wallet.FromJSON(val)
		})
	})

	if err == ErrNotFound {
		// If wallet doesn't exist, create a new one with zero balance
		wallet.Balance = 0
		return wallet, nil
	}

	return wallet, err
}

// GetWalletBalance retrieves the balance of a wallet by wallet ID
func (d *DB) GetWalletBalance(walletID string) (float64, error) {
	wallet, err := d.GetWallet(walletID)
	if err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}

// SaveWallet saves a wallet to the database
func (d *DB) SaveWallet(wallet *models.Wallet) error {
	data, err := wallet.ToJSON()
	if err != nil {
		return err
	}

	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Set(wallet.Key(), data)
	})
}

// SaveTransaction saves a transaction to the database
func (d *DB) SaveTransaction(tx *models.Transaction) error {
	data, err := tx.ToJSON()
	if err != nil {
		return err
	}

	return d.db.Update(func(txn *badger.Txn) error {
		// Save by transaction ID
		if err := txn.Set(tx.Key(), data); err != nil {
			return err
		}

		// Save by wallet ID (for indexing)
		return txn.Set(tx.WalletKey(), data)
	})
}

// GetTransactionsByWallet retrieves transactions for a wallet with pagination
func (d *DB) GetTransactionsByWallet(walletID string, limit, offset int) ([]*models.Transaction, error) {
	prefix := []byte("wallet:" + walletID + ":transaction:")
	var transactions []*models.Transaction

	err := d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = limit
		opts.Prefix = prefix

		it := txn.NewIterator(opts)
		defer it.Close()

		// Skip to offset
		i := 0
		for it.Seek(prefix); it.Valid() && i < offset; it.Next() {
			i++
		}

		// Collect transactions up to limit
		i = 0
		for ; it.Valid() && i < limit; it.Next() {
			item := it.Item()

			var tx models.Transaction
			err := item.Value(func(val []byte) error {
				return tx.FromJSON(val)
			})

			if err != nil {
				return err
			}

			transactions = append(transactions, &tx)
			i++
		}

		return nil
	})

	return transactions, err
}

// AddCurrency adds currency to a wallet and records the transaction
func (d *DB) AddCurrency(walletID string, amount float64, description string, additionalData map[string]interface{}) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	tx := &models.Transaction{
		ID:            generateID(),
		WalletID:      walletID,
		Amount:        amount,
		Description:   description,
		AdditionalData: additionalData,
		Timestamp:     time.Now(),
	}

	err := d.db.Update(func(txn *badger.Txn) error {
		// Get wallet
		wallet := &models.Wallet{WalletID: walletID}
		item, err := txn.Get(wallet.Key())

		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}

		if err == nil {
			// Wallet exists, read it
			err = item.Value(func(val []byte) error {
				return wallet.FromJSON(val)
			})

			if err != nil {
				return err
			}
		} else {
			// Wallet doesn't exist, initialize with zero balance
			wallet.Balance = 0
		}

		// Update wallet balance
		wallet.Balance += amount

		// Save wallet
		walletData, err := wallet.ToJSON()
		if err != nil {
			return err
		}

		if err := txn.Set(wallet.Key(), walletData); err != nil {
			return err
		}

		// Save transaction
		txData, err := tx.ToJSON()
		if err != nil {
			return err
		}

		if err := txn.Set(tx.Key(), txData); err != nil {
			return err
		}

		// Save transaction by wallet ID (for indexing)
		return txn.Set(tx.WalletKey(), txData)
	})

	if err != nil {
		return nil, err
	}

	return tx, nil
}

// RemoveCurrency removes currency from a wallet and records the transaction
func (d *DB) RemoveCurrency(walletID string, amount float64, description string, additionalData map[string]interface{}) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	tx := &models.Transaction{
		ID:            generateID(),
		WalletID:      walletID,
		Amount:        -amount, // Negative amount for removal
		Description:   description,
		AdditionalData: additionalData,
		Timestamp:     time.Now(),
	}

	err := d.db.Update(func(txn *badger.Txn) error {
		// Get wallet
		wallet := &models.Wallet{WalletID: walletID}
		item, err := txn.Get(wallet.Key())

		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrInsufficientFunds
			}
			return err
		}

		// Wallet exists, read it
		err = item.Value(func(val []byte) error {
			return wallet.FromJSON(val)
		})

		if err != nil {
			return err
		}

		// Check if wallet has enough balance
		if wallet.Balance < amount {
			return ErrInsufficientFunds
		}

		// Update wallet balance
		wallet.Balance -= amount

		// Save wallet
		walletData, err := wallet.ToJSON()
		if err != nil {
			return err
		}

		if err := txn.Set(wallet.Key(), walletData); err != nil {
			return err
		}

		// Save transaction
		txData, err := tx.ToJSON()
		if err != nil {
			return err
		}

		if err := txn.Set(tx.Key(), txData); err != nil {
			return err
		}

		// Save transaction by wallet ID (for indexing)
		return txn.Set(tx.WalletKey(), txData)
	})

	if err != nil {
		return nil, err
	}

	return tx, nil
}

// RunGC runs garbage collection on the database
func (d *DB) RunGC() error {
	return d.db.RunValueLogGC(0.5)
}

// generateID generates a unique ID for transactions
func generateID() string {
	return filepath.Base(time.Now().Format("20060102150405.000000000"))
}