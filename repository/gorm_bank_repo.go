package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gotest/config"
	"gotest/domain"
)

type gormBankRepository struct {
	db *gorm.DB
}

// NewGORMBankRepository membuat repository berbasis GORM Database
func NewGORMBankRepository(db *gorm.DB) domain.BankRepository {
	return &gormBankRepository{db: db}
}

// InitDB menginisialisasi koneksi Database (Dukungan Multi-Driver: SQLite & PostgreSQL)
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	if cfg.DBDriver == "postgres" {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
		dialector = postgres.Open(dsn)
	} else {
		dialector = sqlite.Open(cfg.DBPath)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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

func (r *gormBankRepository) GetAccountByNumber(ctx context.Context, accNum string) (*domain.Account, error) {
	var acc domain.Account
	err := r.db.WithContext(ctx).First(&acc, "account_number = ?", accNum).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrAccountNotFound
	}
	return &acc, err
}

func (r *gormBankRepository) GetTransactionsByAccountID(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	var txs []domain.Transaction
	err := r.db.WithContext(ctx).
		Where("from_account_id = ? OR to_account_id = ?", accountID, accountID).
		Order("timestamp desc").
		Find(&txs).Error
	return txs, err
}
