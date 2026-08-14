package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"gotest/delivery"
	"gotest/repository"
	"gotest/usecase"
)

func main() {
	// 1. Inisialisasi Database SQLite Nyata ('bank.db')
	db, err := repository.InitDB("bank.db")
	if err != nil {
		log.Fatalf("Gagal terhubung ke Database: %v", err)
	}
	fmt.Println("💾 Database SQLite ('bank.db') & Auto-Migration Berhasil!")

	// 2. Secret Key JWT
	jwtSecret := "SuperSecretBankKey2026"

	// 3. Inisialisasi GORM Repositories
	bankRepo := repository.NewGORMBankRepository(db)
	userRepo := repository.NewGORMUserRepository(db)

	// 4. Inisialisasi Usecases
	bankUsecase := usecase.NewBankUsecase(bankRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, bankUsecase, jwtSecret)

	// 5. Inisialisasi Gin Handlers & Middlewares
	ginBankHandler := delivery.NewGinBankHandler(bankUsecase, authUsecase)
	ginAuthMiddleware := delivery.NewGinAuthMiddleware(authUsecase)

	// 6. Inisialisasi Gin Engine (Termasuk Logger & Recovery Middleware)
	r := gin.Default()

	// Pasang CORS Middleware (Agar bisa diakses oleh Frontend React/Vue)
	r.Use(delivery.CORSMiddleware())

	// -------------------------------------------------------------------------
	// GIN ROUTE GROUPS
	// -------------------------------------------------------------------------

	// Public Routes Group (/auth)
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", ginBankHandler.Register)
		authGroup.POST("/login", ginBankHandler.Login)
	}

	// Protected Routes Group (/accounts) - Diproteksi JWT Auth Middleware
	accountGroup := r.Group("/accounts")
	accountGroup.Use(ginAuthMiddleware.RequireAuth())
	{
		accountGroup.GET("/get", ginBankHandler.GetAccountByID)
		accountGroup.POST("/transfer", ginBankHandler.Transfer)
	}

	port := ":8080"
	fmt.Printf("\n🚀 High-Performance Gin Engine Bank API berjalan di http://localhost%s\n", port)
	fmt.Println("⚡ Features: Gin Router, CORS, Structured Logger, Panic Recovery, JWT Auth, GORM SQLite DB")

	if err := r.Run(port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
