package main

import (
	"fmt"
	"net/http"

	"gotest/delivery"
	"gotest/repository"
	"gotest/usecase"
)

func main() {
	// 1. Secret key untuk penandatanganan Token JWT
	jwtSecret := "SuperSecretBankKey2026"

	// 2. Inisialisasi Repositories
	bankRepo := repository.NewMemoryBankRepository()
	userRepo := repository.NewMemoryUserRepository()

	// 3. Inisialisasi Usecases
	bankUsecase := usecase.NewBankUsecase(bankRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, bankUsecase, jwtSecret)

	// 4. Inisialisasi Handlers & Middleware
	bankHandler := delivery.NewBankHandler(bankUsecase)
	authHandler := delivery.NewAuthHandler(authUsecase)
	authMiddleware := delivery.NewAuthMiddleware(authUsecase)

	// -------------------------------------------------------------------------
	// REGISTRASI ROUTE HTTP
	// -------------------------------------------------------------------------

	// Public Routes (Bisa diakses siapapun tanpa token)
	http.HandleFunc("/auth/register", authHandler.Register)
	http.HandleFunc("/auth/login", authHandler.Login)

	// Protected Routes (Wajib membawa Header 'Authorization: Bearer <token_jwt>')
	http.HandleFunc("/accounts/get", authMiddleware.Protect(bankHandler.GetAccountByID))
	http.HandleFunc("/accounts/transfer", authMiddleware.Protect(bankHandler.Transfer))

	port := ":8080"
	fmt.Printf("🚀 Enterprise JWT Bank API berjalan di http://localhost%s\n", port)
	fmt.Println("🔒 Protected Routes: /accounts/get & /accounts/transfer (Membutuhkan Bearer Token)")
	fmt.Println("Tekan Ctrl+C untuk menghentikan server.")

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}