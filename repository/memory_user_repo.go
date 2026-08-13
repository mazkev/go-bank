package repository

import (
	"context"
	"sync"

	"gotest/domain"
)

type memoryUserRepository struct {
	mu              sync.RWMutex
	users           map[string]*domain.User
	usersByUsername map[string]*domain.User
}

// NewMemoryUserRepository mengembalikan instance implementasi UserRepository
func NewMemoryUserRepository() domain.UserRepository {
	return &memoryUserRepository{
		users:           make(map[string]*domain.User),
		usersByUsername: make(map[string]*domain.User),
	}
}

func (r *memoryUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByUsername[user.Username]; exists {
		return domain.ErrUserAlreadyExists
	}

	r.users[user.ID] = user
	r.usersByUsername[user.Username] = user
	return nil
}

func (r *memoryUserRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.usersByUsername[username]
	if !exists {
		return nil, domain.ErrInvalidCredentials
	}
	return user, nil
}

func (r *memoryUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, domain.ErrInvalidCredentials
	}
	return user, nil
}
