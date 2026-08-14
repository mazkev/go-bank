package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"gotest/domain"
)

type bankUsecase struct {
	repo domain.BankRepository
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

	// Generate ID unik berbasis UUID (contoh: ACC-A1B2C3D4)
	shortUUID := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", "")[:8])
	newAccID := fmt.Sprintf("ACC-%s", shortUUID)

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

	var createdTx *domain.Transaction

	// ExecTx membungkus 5 langkah berikut ke dalam 1 DB Transaction (ACID)
	err := u.repo.ExecTx(ctx, func(txRepo domain.BankRepository) error {
		// 1. Ambil data rekening pengirim
		fromAcc, err := txRepo.GetAccountByID(ctx, fromID)
		if err != nil {
			return err
		}

		// 2. Ambil data rekening penerima
		toAcc, err := txRepo.GetAccountByID(ctx, toID)
		if err != nil {
			return err
		}

		// 3. Validasi Kecukupan Saldo (Business Rule)
		if fromAcc.Balance < amount {
			return domain.ErrInsufficientBalance
		}

		// 4. Update Saldo Pengirim & Penerima
		if err := txRepo.UpdateAccountBalance(ctx, fromID, fromAcc.Balance-amount); err != nil {
			return err
		}
		if err := txRepo.UpdateAccountBalance(ctx, toID, toAcc.Balance+amount); err != nil {
			return err
		}

		// 5. Catat Histori Transaksi (UUID Unik)
		shortTxUUID := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", "")[:8])
		newTxID := fmt.Sprintf("TX-%s", shortTxUUID)

		tx := &domain.Transaction{
			ID:            newTxID,
			FromAccountID: fromID,
			ToAccountID:   toID,
			Amount:        amount,
			Type:          "TRANSFER",
			Timestamp:     time.Now(),
		}

		if err := txRepo.CreateTransaction(ctx, tx); err != nil {
			return err
		}

		createdTx = tx
		return nil // Mengembalikan nil = Commit Transaksi Otomatis
	})

	// Jika ada error sekecil apapun di dalam ExecTx = Rollback Transaksi Otomatis!
	if err != nil {
		return nil, err
	}

	return createdTx, nil
}

