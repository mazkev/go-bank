package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"gotest/domain"
)

type bankUsecase struct {
	repo domain.BankRepository
}

func NewBankUsecase(repo domain.BankRepository) domain.BankUsecase {
	return &bankUsecase{repo: repo}
}

func (u *bankUsecase) CreateAccount(ctx context.Context, ownerName, accountType, pin string, initialBalance float64) (*domain.Account, error) {
	if initialBalance < 0 {
		return nil, domain.ErrInvalidAmount
	}
	if len(pin) != 6 {
		return nil, fmt.Errorf("PIN transaksi harus terdiri dari 6-digit angka")
	}

	// 1. Hash PIN Transaksi 6-digit dengan Bcrypt
	hashedPIN, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 2. Generasi Nomor Rekening 10-Digit Unik (Format: 10 + 8 digit acak)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	accNumber := fmt.Sprintf("10%08d", rng.Intn(100000000))

	// 3. Generasi ID Unik UUID
	shortUUID := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", "")[:8])
	newAccID := fmt.Sprintf("ACC-%s", shortUUID)

	if accountType == "" {
		accountType = "SAVINGS"
	}

	acc := &domain.Account{
		ID:            newAccID,
		AccountNumber: accNumber,
		OwnerName:     ownerName,
		AccountType:   accountType,
		Currency:      "IDR",
		Balance:       initialBalance,
		PIN:           string(hashedPIN),
		Status:        "ACTIVE",
		CreatedAt:     time.Now(),
	}

	if err := u.repo.CreateAccount(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func (u *bankUsecase) GetAccountByID(ctx context.Context, id string) (*domain.Account, error) {
	return u.repo.GetAccountByID(ctx, id)
}

func (u *bankUsecase) Transfer(ctx context.Context, fromID, toID, pin string, amount float64, description string) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if fromID == toID {
		return nil, domain.ErrSameAccountTransfer
	}

	var createdTx *domain.Transaction

	err := u.repo.ExecTx(ctx, func(txRepo domain.BankRepository) error {
		fromAcc, err := txRepo.GetAccountByID(ctx, fromID)
		if err != nil {
			return err
		}

		if fromAcc.Status != "ACTIVE" {
			return domain.ErrAccountFrozen
		}

		// Verifikasi PIN Transaksi 6-Digit Pengirim
		if err := bcrypt.CompareHashAndPassword([]byte(fromAcc.PIN), []byte(pin)); err != nil {
			return domain.ErrInvalidPIN
		}

		toAcc, err := txRepo.GetAccountByID(ctx, toID)
		if err != nil {
			return err
		}

		if fromAcc.Balance < amount {
			return domain.ErrInsufficientBalance
		}

		// Update Saldo
		if err := txRepo.UpdateAccountBalance(ctx, fromID, fromAcc.Balance-amount); err != nil {
			return err
		}
		if err := txRepo.UpdateAccountBalance(ctx, toID, toAcc.Balance+amount); err != nil {
			return err
		}

		// Buat Histori Transaksi & Nomor Referensi Bank
		shortTxUUID := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", "")[:8])
		newTxID := fmt.Sprintf("TX-%s", shortTxUUID)
		refNum := fmt.Sprintf("REF/%s/%s", time.Now().Format("20060102"), shortTxUUID)

		if description == "" {
			description = fmt.Sprintf("Transfer ke %s (%s)", toAcc.OwnerName, toAcc.AccountNumber)
		}

		tx := &domain.Transaction{
			ID:            newTxID,
			RefNumber:     refNum,
			FromAccountID: fromID,
			ToAccountID:   toID,
			Amount:        amount,
			Type:          "TRANSFER",
			Description:   description,
			Timestamp:     time.Now(),
		}

		if err := txRepo.CreateTransaction(ctx, tx); err != nil {
			return err
		}

		createdTx = tx
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdTx, nil
}

func (u *bankUsecase) Withdraw(ctx context.Context, accountID, pin string, amount float64) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	var createdTx *domain.Transaction

	err := u.repo.ExecTx(ctx, func(txRepo domain.BankRepository) error {
		acc, err := txRepo.GetAccountByID(ctx, accountID)
		if err != nil {
			return err
		}

		if acc.Status != "ACTIVE" {
			return domain.ErrAccountFrozen
		}

		// Verifikasi PIN 6-digit
		if err := bcrypt.CompareHashAndPassword([]byte(acc.PIN), []byte(pin)); err != nil {
			return domain.ErrInvalidPIN
		}

		if acc.Balance < amount {
			return domain.ErrInsufficientBalance
		}

		if err := txRepo.UpdateAccountBalance(ctx, accountID, acc.Balance-amount); err != nil {
			return err
		}

		shortTxUUID := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", "")[:8])
		refNum := fmt.Sprintf("REF/WDR/%s/%s", time.Now().Format("20060102"), shortTxUUID)

		tx := &domain.Transaction{
			ID:            fmt.Sprintf("TX-%s", shortTxUUID),
			RefNumber:     refNum,
			FromAccountID: accountID,
			Amount:        amount,
			Type:          "WITHDRAWAL",
			Description:   "Tarik Tunai di ATM / Counter",
			Timestamp:     time.Now(),
		}

		if err := txRepo.CreateTransaction(ctx, tx); err != nil {
			return err
		}

		createdTx = tx
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdTx, nil
}

func (u *bankUsecase) GetMutations(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	return u.repo.GetTransactionsByAccountID(ctx, accountID)
}
