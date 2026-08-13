package delivery

import (
	"encoding/json"
	"net/http"

	"gotest/domain"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

func NewAuthHandler(u domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: u}
}

type RegisterRequest struct {
	Username       string  `json:"username"`
	Password       string  `json:"password"`
	InitialBalance float64 `json:"initial_balance"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Message string       `json:"message"`
	Token   string       `json:"token"`
	User    *domain.User `json:"user"`
}

// POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	user, token, err := h.authUsecase.Register(r.Context(), req.Username, req.Password, req.InitialBalance)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{
		Message: "Registrasi akun & rekening berhasil",
		Token:   token,
		User:    user,
	})
}

// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	token, user, err := h.authUsecase.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		Message: "Login berhasil",
		Token:   token,
		User:    user,
	})
}
