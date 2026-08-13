package repository

import (
	"context"
	"sync"

	"gotest/domain"
)

type memoryBankRepository struct {
	mu           sync.RWMutex
	accounts     map[string]*domain.Account
	transactions []domain.Transaction
}

// NewMemoryBankRepository mengembalikan instance implementasi BankRepository
func NewMemoryBankRepository() domain.BankRepository {
	return &memoryBankRepository{
		accounts:     make(map[string]*domain.Account),
		transactions: make([]domain.Transaction, 0),
	}
}

func (r *memoryBankRepository) CreateAccount(ctx context.Context, acc *domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.accounts[acc.ID] = acc
	return nil
}

func (r *memoryBankRepository) GetAccountByID(ctx context.Context, id string) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	acc, exists := r.accounts[id]
	if !exists {
		return nil, domain.ErrAccountNotFound
	}
	return acc, nil
}

func (r *memoryBankRepository) UpdateAccountBalance(ctx context.Context, id string, newBalance float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acc, exists := r.accounts[id]
	if !exists {
		return domain.ErrAccountNotFound
	}

	acc.Balance = newBalance
	return nil
}

func (r *memoryBankRepository) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactions = append(r.transactions, *tx)
	return nil
}
