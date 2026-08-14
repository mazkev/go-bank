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

// Struct transaksi khusus untuk ExecTx agar tidak terjadi RLock/Lock deadlock pada goroutine yang sama
type txMemoryRepository struct {
	repo *memoryBankRepository
}

func (r *txMemoryRepository) CreateAccount(ctx context.Context, acc *domain.Account) error {
	r.repo.accounts[acc.ID] = acc
	return nil
}

func (r *txMemoryRepository) GetAccountByID(ctx context.Context, id string) (*domain.Account, error) {
	acc, exists := r.repo.accounts[id]
	if !exists {
		return nil, domain.ErrAccountNotFound
	}
	return acc, nil
}

func (r *txMemoryRepository) UpdateAccountBalance(ctx context.Context, id string, newBalance float64) error {
	acc, exists := r.repo.accounts[id]
	if !exists {
		return domain.ErrAccountNotFound
	}
	acc.Balance = newBalance
	return nil
}

func (r *txMemoryRepository) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	r.repo.transactions = append(r.repo.transactions, *tx)
	return nil
}

func (r *txMemoryRepository) ExecTx(ctx context.Context, fn func(repo domain.BankRepository) error) error {
	return fn(r)
}

// ExecTx untuk in-memory repository (mengisolasi lock dalam 1 transaksi)
func (r *memoryBankRepository) ExecTx(ctx context.Context, fn func(repo domain.BankRepository) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return fn(&txMemoryRepository{repo: r})
}

func (r *memoryBankRepository) GetAccountByNumber(ctx context.Context, accNum string) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, acc := range r.accounts {
		if acc.AccountNumber == accNum {
			return acc, nil
		}
	}
	return nil, domain.ErrAccountNotFound
}

func (r *memoryBankRepository) GetTransactionsByAccountID(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Transaction
	for _, tx := range r.transactions {
		if tx.FromAccountID == accountID || tx.ToAccountID == accountID {
			result = append(result, tx)
		}
	}
	return result, nil
}

// Dan tambahkan juga di struct txMemoryRepository (di bagian paling bawah file memory_bank_repo.go):
func (r *txMemoryRepository) GetAccountByNumber(ctx context.Context, accNum string) (*domain.Account, error) {
	for _, acc := range r.repo.accounts {
		if acc.AccountNumber == accNum {
			return acc, nil
		}
	}
	return nil, domain.ErrAccountNotFound
}

func (r *txMemoryRepository) GetTransactionsByAccountID(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	var result []domain.Transaction
	for _, tx := range r.repo.transactions {
		if tx.FromAccountID == accountID || tx.ToAccountID == accountID {
			result = append(result, tx)
		}
	}
	return result, nil
}
