package delivery

import (
	"context"
	"net/http"
	"strings"

	"gotest/domain"
)

type contextKey string

const UserClaimsKey contextKey = "user_claims"

type AuthMiddleware struct {
	authUsecase domain.AuthUsecase
}

func NewAuthMiddleware(u domain.AuthUsecase) *AuthMiddleware {
	return &AuthMiddleware{authUsecase: u}
}

// Protect adalah wrapper HTTP handler yang memeriksa token JWT
func (m *AuthMiddleware) Protect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "Header Authorization diperlukan")
			return
		}

		// Header harus berformat: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeError(w, http.StatusUnauthorized, "Format header harus 'Bearer <token>'")
			return
		}

		tokenString := parts[1]
		claims, err := m.authUsecase.ValidateToken(tokenString)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}

		// Sisipkan JWT claims ke dalam Request Context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

// Helper untuk mengambil JWT Claims dari Request Context
func GetUserClaimsFromContext(ctx context.Context) (*domain.JWTClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*domain.JWTClaims)
	return claims, ok
}
