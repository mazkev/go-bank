package main

import (
	"fmt"
	"log"
	"net/http"

	"gotest/delivery"
	"gotest/repository"
	"gotest/usecase"
)

func main() {
	// 1. Inisialisasi Database SQLite Nyata ('bank.db') & Auto-Migrate Tabel
	db, err := repository.InitDB("bank.db")
	if err != nil {
		log.Fatalf("Gagal terhubung ke Database: %v", err)
	}
	fmt.Println("💾 Database SQLite ('bank.db') & Auto-Migration Berhasil!")

	// 2. Secret Key JWT
	jwtSecret := "SuperSecretBankKey2026"

	// 3. Inisialisasi GORM Repositories (Menggantikan In-Memory)
	bankRepo := repository.NewGORMBankRepository(db)
	userRepo := repository.NewGORMUserRepository(db)

	// 4. Inisialisasi Usecases (Menggunakan Dependency Injection ke GORM Repo)
	bankUsecase := usecase.NewBankUsecase(bankRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, bankUsecase, jwtSecret)

	// 5. Inisialisasi Handlers & Middleware
	bankHandler := delivery.NewBankHandler(bankUsecase)
	authHandler := delivery.NewAuthHandler(authUsecase)
	authMiddleware := delivery.NewAuthMiddleware(authUsecase)

	// -------------------------------------------------------------------------
	// REGISTRASI ROUTE HTTP
	// -------------------------------------------------------------------------
	http.HandleFunc("/auth/register", authHandler.Register)
	http.HandleFunc("/auth/login", authHandler.Login)
	http.HandleFunc("/accounts/get", authMiddleware.Protect(bankHandler.GetAccountByID))
	http.HandleFunc("/accounts/transfer", authMiddleware.Protect(bankHandler.Transfer))

	port := ":8080"
	fmt.Printf("🚀 Enterprise GORM SQLite Bank API berjalan di http://localhost%s\n", port)
	fmt.Println("🔒 Protected Routes: /accounts/get & /accounts/transfer")
	fmt.Println("Tekan Ctrl+C untuk menghentikan server.")

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
