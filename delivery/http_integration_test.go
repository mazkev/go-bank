package delivery_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest/delivery"
	"gotest/repository"
	"gotest/usecase"
)

// Helper setup HTTP Test Server dengan Clean Architecture & Auth Middleware
func setupTestServer() *httptest.Server {
	jwtSecret := "TestSecretKey123"

	bankRepo := repository.NewMemoryBankRepository()
	userRepo := repository.NewMemoryUserRepository()

	bankUsecase := usecase.NewBankUsecase(bankRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, bankUsecase, jwtSecret)

	bankHandler := delivery.NewBankHandler(bankUsecase)
	authHandler := delivery.NewAuthHandler(authUsecase)
	authMiddleware := delivery.NewAuthMiddleware(authUsecase)

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/accounts/get", authMiddleware.Protect(bankHandler.GetAccountByID))
	mux.HandleFunc("/accounts/transfer", authMiddleware.Protect(bankHandler.Transfer))

	return httptest.NewServer(mux)
}

func TestHTTPIntegration_FullFlow(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	client := ts.Client()

	// 1. TEST: Akses Endpoint Tanpa Token (Harus Status 401 Unauthorized)
	t.Run("1. Unauthorized Access Without Token", func(t *testing.T) {
		resp, err := client.Get(ts.URL + "/accounts/get?id=ACC-001")
		if err != nil {
			t.Fatalf("Request gagal: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Ekspektasi 401 Unauthorized, tapi dapat: %d", resp.StatusCode)
		}
	})

	var tokenBudi string
	var accountIDBudi string
	var accountIDSiti string

	// 2. TEST: Register User 'budi' (Dapat Akun Saldo 1.000.000)
	t.Run("2. Register User Budi", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"username":        "budi",
			"password":        "password123",
			"initial_balance": 1000000,
		})

		resp, err := client.Post(ts.URL+"/auth/register", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Register gagal: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Ekspektasi Status 201 Created, dapat: %d", resp.StatusCode)
		}

		var res delivery.AuthResponse
		json.NewDecoder(resp.Body).Decode(&res)

		if res.Token == "" {
			t.Errorf("Token JWT tidak boleh kosong!")
		}
		tokenBudi = res.Token
		accountIDBudi = res.User.AccountID
		t.Logf("Token Budi: %s..., AccountID: %s", tokenBudi[:20], accountIDBudi)
	})

	// 3. TEST: Register User 'siti' (Dapat Akun Saldo 500.000)
	t.Run("3. Register User Siti", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"username":        "siti",
			"password":        "password456",
			"initial_balance": 500000,
		})

		resp, err := client.Post(ts.URL+"/auth/register", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Register gagal: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Ekspektasi Status 201 Created, dapat: %d", resp.StatusCode)
		}

		var res delivery.AuthResponse
		json.NewDecoder(resp.Body).Decode(&res)
		accountIDSiti = res.User.AccountID
	})

	// 4. TEST: Login User Budi dengan Password Salah (Harus Status 401)
	t.Run("4. Login Wrong Password", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{
			"username": "budi",
			"password": "wrongpassword",
		})

		resp, err := client.Post(ts.URL+"/auth/login", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Login gagal: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Ekspektasi Status 401 Unauthorized, dapat: %d", resp.StatusCode)
		}
	})

	// 5. TEST: Get Account Details Membawa Bearer Token
	t.Run("5. Protected Get Account Details With JWT Token", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/accounts/get?id="+accountIDBudi, nil)
		req.Header.Set("Authorization", "Bearer "+tokenBudi)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Get account gagal: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Ekspektasi Status 200 OK, dapat: %d", resp.StatusCode)
		}

		var acc map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&acc)

		if acc["id"] != accountIDBudi || acc["owner_name"] != "budi" {
			t.Errorf("Data akun tidak sesuai: %v", acc)
		}
	})

	// 6. TEST: Transfer Saldo Rp 300.000 dari Budi ke Siti Membawa Bearer Token
	t.Run("6. Protected Transfer Saldo With JWT Token", func(t *testing.T) {
		transferBody, _ := json.Marshal(map[string]interface{}{
			"from_account_id": accountIDBudi,
			"to_account_id":   accountIDSiti,
			"amount":          300000,
		})

		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/accounts/transfer", bytes.NewBuffer(transferBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenBudi)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Transfer gagal: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Ekspektasi Status 200 OK, dapat: %d", resp.StatusCode)
		}

		var tx map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&tx)

		if tx["amount"] != float64(300000) || tx["type"] != "TRANSFER" {
			t.Errorf("Transaksi tidak sesuai: %v", tx)
		}
	})
}
