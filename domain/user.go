package domain

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Custom Auth Errors
var (
	ErrUserAlreadyExists  = errors.New("username sudah terdaftar")
	ErrInvalidCredentials = errors.New("username atau password salah")
	ErrUnauthorized       = errors.New("akses ditolak: token JWT tidak valid atau kadaluarsa")
)

// Entity User
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Jangan tampilkan password di JSON output!
	AccountID string    `json:"account_id"`
	CreatedAt time.Time `json:"created_at"`
}

// JWT Claims Struct
type JWTClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	AccountID string `json:"account_id"`
	jwt.RegisteredClaims
}

// Interfaces
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
}

type AuthUsecase interface {
	Register(ctx context.Context, username, password string, initialBalance float64) (*User, string, error)
	Login(ctx context.Context, username, password string) (string, *User, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
}
