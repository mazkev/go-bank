package http

import (
	"encoding/json"
	"net/http"

	"gotest/domain"
)

type BankHandler struct {
	usecase domain.BankUsecase
}

// NewBankHandler menghubungkan HTTP Layer dengan Usecase Layer (Dependency Injection)
func NewBankHandler(u domain.BankUsecase) *BankHandler {
	return &BankHandler{
		usecase: u,
	}
}

// Helper Response JSON
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// DTO Requests
type CreateAccountRequest struct {
	OwnerName      string  `json:"owner_name"`
	InitialBalance float64 `json:"initial_balance"`
}

type TransferRequest struct {
	FromAccountID string  `json:"from_account_id"`
	ToAccountID   string  `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}

// POST /accounts/create
func (h *BankHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
		return
	}

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	// Teruskan request context r.Context() ke usecase
	acc, err := h.usecase.CreateAccount(r.Context(), req.OwnerName, req.InitialBalance)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, acc)
}

// GET /accounts/get?id=ACC-001
func (h *BankHandler) GetAccountByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Parameter 'id' wajib diisi")
		return
	}

	acc, err := h.usecase.GetAccountByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, acc)
}

// POST /accounts/transfer
func (h *BankHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method tidak diizinkan")
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	tx, err := h.usecase.Transfer(r.Context(), req.FromAccountID, req.ToAccountID, req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, tx)
}
