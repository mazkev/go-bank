package delivery

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gotest/domain"
)

type GinAuthMiddleware struct {
	authUsecase domain.AuthUsecase
}

func NewGinAuthMiddleware(u domain.AuthUsecase) *GinAuthMiddleware {
	return &GinAuthMiddleware{authUsecase: u}
}

// RequireAuth adalah Gin Middleware untuk memproteksi Endpoint
func (m *GinAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Header Authorization diperlukan"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Format header harus 'Bearer <token>'"})
			c.Abort()
			return
		}

		claims, err := m.authUsecase.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Simpan JWT Claims ke dalam Gin Context
		c.Set("user_claims", claims)
		c.Next()
	}
}

// CORSMiddleware mengizinkan akses dari Frontend (React/Vue/Angular)
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
