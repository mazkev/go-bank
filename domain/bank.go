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
)

// Entities
type Account struct {
	ID        string    `json:"id"`
	OwnerName string    `json:"owner_name"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type Transaction struct {
	ID            string    `json:"id"`
	FromAccountID string    `json:"from_account_id,omitempty"`
	ToAccountID   string    `json:"to_account_id,omitempty"`
	Amount        float64   `json:"amount"`
	Type          string    `json:"type"`
	Timestamp     time.Time `json:"timestamp"`
}

// -----------------------------------------------------------------------------
// INTERFACES (Kontrak Arsitektur Clean Architecture)
// -----------------------------------------------------------------------------

// BankRepository adalah kontrak untuk Data Access Layer (DB / Memory)
type BankRepository interface {
	CreateAccount(ctx context.Context, acc *Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	UpdateAccountBalance(ctx context.Context, id string, newBalance float64) error
	CreateTransaction(ctx context.Context, tx *Transaction) error
}

// BankUsecase adalah kontrak untuk Business Logic Layer
type BankUsecase interface {
	CreateAccount(ctx context.Context, ownerName string, initialBalance float64) (*Account, error)
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	Transfer(ctx context.Context, fromID, toID string, amount float64) (*Transaction, error)
}
