package repository

import (
	"context"
	"errors"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"gotest/domain"
)

type gormBankRepository struct {
	db *gorm.DB
}

// NewGORMBankRepository membuat repository berbasis GORM Database
func NewGORMBankRepository(db *gorm.DB) domain.BankRepository {
	return &gormBankRepository{db: db}
}

// Inisialisasi Database SQLite & Auto-Migrate Tabel Domain
func InitDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-Migrate: Membuat tabel database otomatis dari Struct Domain
	err = db.AutoMigrate(&domain.Account{}, &domain.User{}, &domain.Transaction{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *gormBankRepository) CreateAccount(ctx context.Context, acc *domain.Account) error {
	return r.db.WithContext(ctx).Create(acc).Error
}

func (r *gormBankRepository) GetAccountByID(ctx context.Context, id string) (*domain.Account, error) {
	var acc domain.Account
	err := r.db.WithContext(ctx).First(&acc, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrAccountNotFound
	}
	return &acc, err
}

func (r *gormBankRepository) UpdateAccountBalance(ctx context.Context, id string, newBalance float64) error {
	result := r.db.WithContext(ctx).Model(&domain.Account{}).Where("id = ?", id).Update("balance", newBalance)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *gormBankRepository) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// ExecTx mengeksekusi sekumpulan operasi DB di dalam 1 GORM Transaction (ACID)
func (r *gormBankRepository) ExecTx(ctx context.Context, fn func(repo domain.BankRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &gormBankRepository{db: tx}
		return fn(txRepo)
	})
}
