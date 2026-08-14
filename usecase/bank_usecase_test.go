package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

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

func (m *MockBankRepository) GetAccountByNumber(ctx context.Context, accNum string) (*domain.Account, error) {
	args := m.Called(ctx, accNum)
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

func (m *MockBankRepository) GetTransactionsByAccountID(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockBankRepository) ExecTx(ctx context.Context, fn func(repo domain.BankRepository) error) error {
	return fn(m)
}

// -----------------------------------------------------------------------------
// UNIT TEST SUITE UNTUK LOGIKA BANK USECASE NYATA
// -----------------------------------------------------------------------------

// Test 1: Transfer Saldo Sukses dengan Verifikasi PIN 6-Digit
func TestTransfer_Success(t *testing.T) {
	mockRepo := new(MockBankRepository)
	uc := usecase.NewBankUsecase(mockRepo)
	ctx := context.Background()

	hashedPIN, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	fromAcc := &domain.Account{
		ID:            "ACC-001",
		AccountNumber: "1029384756",
		OwnerName:     "Budi",
		Balance:       1000000,
		PIN:           string(hashedPIN),
		Status:        "ACTIVE",
	}
	toAcc := &domain.Account{
		ID:            "ACC-002",
		AccountNumber: "1098765432",
		OwnerName:     "Siti",
		Balance:       500000,
		PIN:           string(hashedPIN),
		Status:        "ACTIVE",
	}

	mockRepo.On("GetAccountByID", ctx, "ACC-001").Return(fromAcc, nil)
	mockRepo.On("GetAccountByID", ctx, "ACC-002").Return(toAcc, nil)
	mockRepo.On("UpdateAccountBalance", ctx, "ACC-001", float64(700000)).Return(nil)
	mockRepo.On("UpdateAccountBalance", ctx, "ACC-002", float64(800000)).Return(nil)
	mockRepo.On("CreateTransaction", ctx, mock.AnythingOfType("*domain.Transaction")).Return(nil)

	tx, err := uc.Transfer(ctx, "ACC-001", "ACC-002", "123456", 300000, "Bayar Kopi")

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, float64(300000), tx.Amount)
	assert.Equal(t, "TRANSFER", tx.Type)
	assert.Equal(t, "Bayar Kopi", tx.Description)

	mockRepo.AssertExpectations(t)
}

// Test 2: Transfer Gagal karena PIN 6-Digit Salah
func TestTransfer_InvalidPIN(t *testing.T) {
	mockRepo := new(MockBankRepository)
	uc := usecase.NewBankUsecase(mockRepo)
	ctx := context.Background()

	hashedPIN, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	fromAcc := &domain.Account{
		ID:      "ACC-001",
		Balance: 1000000,
		PIN:     string(hashedPIN),
		Status:  "ACTIVE",
	}

	mockRepo.On("GetAccountByID", ctx, "ACC-001").Return(fromAcc, nil)

	tx, err := uc.Transfer(ctx, "ACC-001", "ACC-002", "999999", 100000, "Transfer")

	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.True(t, errors.Is(err, domain.ErrInvalidPIN))

	mockRepo.AssertExpectations(t)
}

// Test 3: Transfer Gagal karena Saldo Pengirim Tidak Cukup
func TestTransfer_InsufficientBalance(t *testing.T) {
	mockRepo := new(MockBankRepository)
	uc := usecase.NewBankUsecase(mockRepo)
	ctx := context.Background()

	hashedPIN, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	fromAcc := &domain.Account{
		ID:      "ACC-001",
		Balance: 100000,
		PIN:     string(hashedPIN),
		Status:  "ACTIVE",
	}
	toAcc := &domain.Account{
		ID:      "ACC-002",
		Balance: 500000,
		Status:  "ACTIVE",
	}

	mockRepo.On("GetAccountByID", ctx, "ACC-001").Return(fromAcc, nil)
	mockRepo.On("GetAccountByID", ctx, "ACC-002").Return(toAcc, nil)

	tx, err := uc.Transfer(ctx, "ACC-001", "ACC-002", "123456", 500000, "Transfer")

	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.True(t, errors.Is(err, domain.ErrInsufficientBalance))

	mockRepo.AssertExpectations(t)
}

// Test 4: Tarik Tunai (Withdrawal) Sukses
func TestWithdraw_Success(t *testing.T) {
	mockRepo := new(MockBankRepository)
	uc := usecase.NewBankUsecase(mockRepo)
	ctx := context.Background()

	hashedPIN, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	acc := &domain.Account{
		ID:      "ACC-001",
		Balance: 500000,
		PIN:     string(hashedPIN),
		Status:  "ACTIVE",
	}

	mockRepo.On("GetAccountByID", ctx, "ACC-001").Return(acc, nil)
	mockRepo.On("UpdateAccountBalance", ctx, "ACC-001", float64(400000)).Return(nil)
	mockRepo.On("CreateTransaction", ctx, mock.AnythingOfType("*domain.Transaction")).Return(nil)

	tx, err := uc.Withdraw(ctx, "ACC-001", "123456", 100000)

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, float64(100000), tx.Amount)
	assert.Equal(t, "WITHDRAWAL", tx.Type)

	mockRepo.AssertExpectations(t)
}
