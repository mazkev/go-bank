package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gotest/domain"
	"gotest/usecase"
)

// -----------------------------------------------------------------------------
// MOCK REPOSITORY IMPLEMENTATION (Objek Tiruan untuk domain.BankRepository)
// -----------------------------------------------------------------------------
type MockBankRepository struct {
	mock.Mock
}

func (m *MockBankRepository) CreateAccount(ctx context.Context, acc *domain.Account) error {
	args := m.Called(ctx, acc)
	return args.Error(0)
}

func (m *MockBankRepository) GetAccountByID(ctx context.Context, id string) (*domain.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockBankRepository) UpdateAccountBalance(ctx context.Context, id string, newBalance float64) error {
	args := m.Called(ctx, id, newBalance)
	return args.Error(0)
}

func (m *MockBankRepository) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockBankRepository) ExecTx(ctx context.Context, fn func(repo domain.BankRepository) error) error {
	// Menjalankan callback fn dengan menggunakan mock repository itu sendiri
	return fn(m)
}

// -----------------------------------------------------------------------------
// UNIT TEST SUITE UNTUK LOGIKA BANK USECASE
// -----------------------------------------------------------------------------

// Test 1: Transfer Saldo Sukses
func TestTransfer_Success(t *testing.T) {
	mockRepo := new(MockBankRepository)
	uc := usecase.NewBankUsecase(mockRepo)
	ctx := context.Background()

	fromAcc := &domain.Account{ID: "ACC-001", OwnerName: "Budi", Balance: 1000000}
	toAcc := &domain.Account{ID: "ACC-002", OwnerName: "Siti", Balance: 500000}

	// Set Ekspektasi Mock:
	mockRepo.On("GetAccountByID", ctx, "ACC-001").Return(fromAcc, nil)
	mockRepo.On("GetAccountByID", ctx, "ACC-002").Return(toAcc, nil)
	mockRepo.On("UpdateAccountBalance", ctx, "ACC-001", float64(700000)).Return(nil)
	mockRepo.On("UpdateAccountBalance", ctx, "ACC-002", float64(800000)).Return(nil)
	mockRepo.On("CreateTransaction", ctx, mock.AnythingOfType("*domain.Transaction")).Return(nil)

	// Eksekusi Usecase
	tx, err := uc.Transfer(ctx, "ACC-001", "ACC-002", 300000)

	// Verifikasi Hasil
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, float64(300000), tx.Amount)
	assert.Equal(t, "TRANSFER", tx.Type)

	// Memastikan semua eksekusi mock sesuai dengan ekspektasi
	mockRepo.AssertExpectations(t)
}

// Test 2: Transfer Gagal karena Saldo Pengirim Tidak Cukup
func TestTransfer_InsufficientBalance(t *testing.T) {
	mockRepo := new(MockBankRepository)
	uc := usecase.NewBankUsecase(mockRepo)
	ctx := context.Background()

	fromAcc := &domain.Account{ID: "ACC-001", OwnerName: "Budi", Balance: 100000}
	toAcc := &domain.Account{ID: "ACC-002", OwnerName: "Siti", Balance: 500000}

	mockRepo.On("GetAccountByID", ctx, "ACC-001").Return(fromAcc, nil)
	mockRepo.On("GetAccountByID", ctx, "ACC-002").Return(toAcc, nil)

	// Eksekusi Usecase: Mencoba transfer 500.000 padahal saldo hanya 100.000
	tx, err := uc.Transfer(ctx, "ACC-001", "ACC-002", 500000)

	// Verifikasi Error
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.True(t, errors.Is(err, domain.ErrInsufficientBalance))

	mockRepo.AssertExpectations(t)
}

// Test 3: Transfer ke Rekening Sendiri (Ditolak)
func TestTransfer_SameAccount(t *testing.T) {
	mockRepo := new(MockBankRepository)
	uc := usecase.NewBankUsecase(mockRepo)
	ctx := context.Background()

	tx, err := uc.Transfer(ctx, "ACC-001", "ACC-001", 100000)

	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.True(t, errors.Is(err, domain.ErrSameAccountTransfer))
}
