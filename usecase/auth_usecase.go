package usecase

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"gotest/domain"
)

type authUsecase struct {
	userRepo    domain.UserRepository
	bankUsecase domain.BankUsecase
	jwtSecret   []byte
	userCounter int64
}

// NewAuthUsecase mengembalikan instance AuthUsecase dengan Dependency Injection
func NewAuthUsecase(uRepo domain.UserRepository, bUsecase domain.BankUsecase, secret string) domain.AuthUsecase {
	return &authUsecase{
		userRepo:    uRepo,
		bankUsecase: bUsecase,
		jwtSecret:   []byte(secret),
	}
}

func (u *authUsecase) Register(ctx context.Context, username, password string, initialBalance float64) (*domain.User, string, error) {
	if len(username) < 3 || len(password) < 6 {
		return nil, "", fmt.Errorf("username minimal 3 karakter dan password minimal 6 karakter")
	}

	// 1. Buat rekening bank otomatis untuk user baru
	acc, err := u.bankUsecase.CreateAccount(ctx, username, initialBalance)
	if err != nil {
		return nil, "", err
	}

	// 2. Hash password dengan Bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// 3. Buat entity User
	newUserID := fmt.Sprintf("USR-%03d", atomic.AddInt64(&u.userCounter, 1))
	user := &domain.User{
		ID:        newUserID,
		Username:  username,
		Password:  string(hashedPassword),
		AccountID: acc.ID,
		CreatedAt: time.Now(),
	}

	if err := u.userRepo.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	// 4. Generate JWT Token
	token, err := u.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (u *authUsecase) Login(ctx context.Context, username, password string) (string, *domain.User, error) {
	user, err := u.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	// Verifikasi Password Hash Bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	// Generate JWT Token
	token, err := u.generateToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (u *authUsecase) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode signing token tidak valid")
		}
		return u.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(*domain.JWTClaims)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	return claims, nil
}

// Helper internal untuk meng-generate JWT Token
func (u *authUsecase) generateToken(user *domain.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token aktif 24 jam

	claims := &domain.JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		AccountID: user.AccountID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(u.jwtSecret)
}
