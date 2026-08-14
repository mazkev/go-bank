package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"gotest/domain"
)

type gormUserRepository struct {
	db *gorm.DB
}

// NewGORMUserRepository membuat repository User berbasis GORM Database
func NewGORMUserRepository(db *gorm.DB) domain.UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	// Cek apakah username sudah dipakai
	var count int64
	r.db.WithContext(ctx).Model(&domain.User{}).Where("username = ?", user.Username).Count(&count)
	if count > 0 {
		return domain.ErrUserAlreadyExists
	}

	return r.db.WithContext(ctx).Create(user).Error
}

func (r *gormUserRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, "username = ?", username).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrInvalidCredentials
	}
	return &user, err
}

func (r *gormUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrInvalidCredentials
	}
	return &user, err
}
