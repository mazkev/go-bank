package usecase

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"gotest/domain"
)

type bankUsecase struct {
	repo       domain.BankRepository
	accCounter int64
	txCounter  int64
}

// NewBankUsecase membuat instance baru BankUsecase dengan Dependency Injection
func NewBankUsecase(repo domain.BankRepository) domain.BankUsecase {
	return &bankUsecase{
		repo: repo,
	}
}

func (u *bankUsecase) CreateAccount(ctx context.Context, ownerName string, initialBalance float64) (*domain.Account, error) {
	if initialBalance < 0 {
		return nil, domain.ErrInvalidAmount
	}

	// Generate ID unik
	newAccID := fmt.Sprintf("ACC-%03d", atomic.AddInt64(&u.accCounter, 1))

	acc := &domain.Account{
		ID:        newAccID,
		OwnerName: ownerName,
		Balance:   initialBalance,
		CreatedAt: time.Now(),
	}

	if err := u.repo.CreateAccount(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func (u *bankUsecase) GetAccountByID(ctx context.Context, id string) (*domain.Account, error) {
	return u.repo.GetAccountByID(ctx, id)
}

func (u *bankUsecase) Transfer(ctx context.Context, fromID, toID string, amount float64) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if fromID == toID {
		return nil, domain.ErrSameAccountTransfer
	}

	// 1. Ambil data rekening pengirim
	fromAcc, err := u.repo.GetAccountByID(ctx, fromID)
	if err != nil {
		return nil, err
	}

	// 2. Ambil data rekening penerima
	toAcc, err := u.repo.GetAccountByID(ctx, toID)
	if err != nil {
		return nil, err
	}

	// 3. Validasi Kecukupan Saldo (Business Rule)
	if fromAcc.Balance < amount {
		return nil, domain.ErrInsufficientBalance
	}

	// 4. Update Saldo Pengirim & Penerima
	if err := u.repo.UpdateAccountBalance(ctx, fromID, fromAcc.Balance-amount); err != nil {
		return nil, err
	}
	if err := u.repo.UpdateAccountBalance(ctx, toID, toAcc.Balance+amount); err != nil {
		return nil, err
	}

	// 5. Catat Histori Transaksi
	newTxID := fmt.Sprintf("TX-%03d", atomic.AddInt64(&u.txCounter, 1))
	tx := &domain.Transaction{
		ID:            newTxID,
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        amount,
		Type:          "TRANSFER",
		Timestamp:     time.Now(),
	}

	if err := u.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	return tx, nil
}
