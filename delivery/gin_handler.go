package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gotest/domain"
)

type GinBankHandler struct {
	bankUsecase domain.BankUsecase
	authUsecase domain.AuthUsecase
}

func NewGinBankHandler(bUsecase domain.BankUsecase, aUsecase domain.AuthUsecase) *GinBankHandler {
	return &GinBankHandler{
		bankUsecase: bUsecase,
		authUsecase: aUsecase,
	}
}

// POST /auth/register
func (h *GinBankHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	user, token, err := h.authUsecase.Register(c.Request.Context(), req.Username, req.Password, req.InitialBalance)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Message: "Registrasi akun & rekening berhasil (Gin Engine)",
		Token:   token,
		User:    user,
	})
}

// POST /auth/login
func (h *GinBankHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	token, user, err := h.authUsecase.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Message: "Login berhasil (Gin Engine)",
		Token:   token,
		User:    user,
	})
}

// GET /accounts/get?id=ACC-001
func (h *GinBankHandler) GetAccountByID(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter 'id' wajib diisi"})
		return
	}

	acc, err := h.bankUsecase.GetAccountByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, acc)
}

// POST /accounts/transfer
func (h *GinBankHandler) Transfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	tx, err := h.bankUsecase.Transfer(c.Request.Context(), req.FromAccountID, req.ToAccountID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}
