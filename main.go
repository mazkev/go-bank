package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"gotest/config"
	"gotest/delivery"
	"gotest/repository"
	"gotest/usecase"
)

func main() {
	// 1. Load Konfigurasi Environment Variables
	cfg := config.LoadConfig()
	gin.SetMode(cfg.GinMode)

	// 2. Inisialisasi Database SQLite
	db, err := repository.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Gagal terhubung ke Database: %v", err)
	}
	fmt.Printf("💾 Database SQLite ('%s') & Auto-Migration Berhasil!\n", cfg.DBPath)

	// 3. Inisialisasi GORM Repositories
	bankRepo := repository.NewGORMBankRepository(db)
	userRepo := repository.NewGORMUserRepository(db)

	// 4. Inisialisasi Usecases
	bankUsecase := usecase.NewBankUsecase(bankRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, bankUsecase, cfg.JWTSecret)

	// 5. Inisialisasi Gin Handlers & Middlewares
	ginBankHandler := delivery.NewGinBankHandler(bankUsecase, authUsecase)
	ginAuthMiddleware := delivery.NewGinAuthMiddleware(authUsecase)

	// 6. Inisialisasi Gin Engine
	r := gin.Default()

	// -------------------------------------------------------------------------
	// GLOBAL MIDDLEWARES: CORS & IP RATE LIMITER
	// -------------------------------------------------------------------------
	r.Use(delivery.CORSMiddleware())
	r.Use(delivery.RateLimitMiddleware(5, 10)) // Batas: 5 req/detik, burst 10

	// -------------------------------------------------------------------------
	// GIN ROUTE GROUPS
	// -------------------------------------------------------------------------
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", ginBankHandler.Register)
		authGroup.POST("/login", ginBankHandler.Login)
	}

	accountGroup := r.Group("/accounts")
	accountGroup.Use(ginAuthMiddleware.RequireAuth())
	{
		accountGroup.GET("/get", ginBankHandler.GetAccountByID)
		accountGroup.POST("/transfer", ginBankHandler.Transfer)
	}

	// -------------------------------------------------------------------------
	// GRACEFUL SHUTDOWN HTTP SERVER SETUP
	// -------------------------------------------------------------------------
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		fmt.Printf("\n🚀 High-Performance Production Bank API berjalan di http://localhost:%s\n", cfg.Port)
		fmt.Println("🛡️  Protection: IP Rate Limiter Active (5 req/sec, burst 10)")
		fmt.Println("⚡ Features: Gin Router, CORS, Graceful Shutdown, .env Config, JWT Auth, GORM SQLite DB")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server Listen error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⚠️  Sinyal pemberhentian diterima! Memulai Graceful Shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Graceful Shutdown terpaksa dihentikan: %v", err)
	}

	log.Println("✅ Server berhasil berhenti secara mulus. Sampai jumpa!")
}
