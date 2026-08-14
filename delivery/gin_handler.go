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

	if req.PIN == "" {
		req.PIN = "123456" // Default 6-digit PIN jika tidak diisi
	}

	user, token, err := h.authUsecase.Register(c.Request.Context(), req.Username, req.Password, req.PIN, req.InitialBalance)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Message: "Registrasi akun & rekening bank berhasil (Gin Engine)",
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

	if req.PIN == "" {
		req.PIN = "123456" // Default 6-digit PIN jika tidak diisi saat pengujian
	}

	tx, err := h.bankUsecase.Transfer(c.Request.Context(), req.FromAccountID, req.ToAccountID, req.PIN, req.Amount, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// POST /accounts/withdraw
func (h *GinBankHandler) Withdraw(c *gin.Context) {
	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	if req.PIN == "" {
		req.PIN = "123456"
	}

	tx, err := h.bankUsecase.Withdraw(c.Request.Context(), req.AccountID, req.PIN, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// GET /accounts/mutations?account_id=ACC-001
func (h *GinBankHandler) GetMutations(c *gin.Context) {
	accountID := c.Query("account_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter 'account_id' wajib diisi"})
		return
	}

	txs, err := h.bankUsecase.GetMutations(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id": accountID,
		"count":      len(txs),
		"mutations":  txs,
	})
}
