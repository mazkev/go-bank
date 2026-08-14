package domain

import (
	"context"
	"errors"
	"time"
)

// Custom Errors
var (
	ErrAccountNotFound     = errors.New("rekening tidak ditemukan")
	ErrInsufficientBalance = errors.New("saldo tidak mencukupi")
	ErrInvalidAmount       = errors.New("jumlah nominal harus lebih besar dari 0")
	ErrSameAccountTransfer = errors.New("tidak bisa mentransfer ke rekening yang sama")
	ErrInvalidPIN          = errors.New("PIN transaksi 6-digit salah")
	ErrAccountFrozen       = errors.New("rekening sedang dibekukan (FROZEN)")
)

// Entity Account (Rekening Bank Nyata)
type Account struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	AccountNumber string    `json:"account_number" gorm:"uniqueIndex"`
	OwnerName     string    `json:"owner_name"`
	AccountType   string    `json:"account_type"` // "SAVINGS", "CHECKING"
	Currency      string    `json:"currency"`     // "IDR", "USD"
	Balance       float64   `json:"balance"`
	PIN           string    `json:"-"`            // Bcrypt Hashed 6-digit PIN
	Status        string    `json:"status"`       // "ACTIVE", "FROZEN", "CLOSED"
	CreatedAt     time.Time `json:"created_at"`
}

// Entity Transaction (Mutasi Rekening Bank Nyata)
type Transaction struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	RefNumber     string    `json:"ref_number" gorm:"uniqueIndex"`
	FromAccountID string    `json:"from_account_id,omitempty"`
	ToAccountID   string    `json:"to_account_id,omitempty"`
	Amount        float64   `json:"amount"`
	Type          string    `json:"type"`        // "DEPOSIT", "WITHDRAWAL", "TRANSFER"
	Description   string    `json:"description"` // Catatan Transaksi
	Timestamp     time.Time `json:"timestamp"`
}

// Kontrak Clean Architecture BankRepository & BankUsecase
type BankRepository interface {
	CreateAccount(ctx context.Context, acc *Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	GetAccountByNumber(ctx context.Context, accNum string) (*Account, error)
	UpdateAccountBalance(ctx context.Context, id string, newBalance float64) error
	CreateTransaction(ctx context.Context, tx *Transaction) error
	GetTransactionsByAccountID(ctx context.Context, accountID string) ([]Transaction, error)
	ExecTx(ctx context.Context, fn func(repo BankRepository) error) error
}

type BankUsecase interface {
	CreateAccount(ctx context.Context, ownerName, accountType, pin string, initialBalance float64) (*Account, error)
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	Transfer(ctx context.Context, fromID, toID, pin string, amount float64, description string) (*Transaction, error)
	Withdraw(ctx context.Context, accountID, pin string, amount float64) (*Transaction, error)
	GetMutations(ctx context.Context, accountID string) ([]Transaction, error)
}
