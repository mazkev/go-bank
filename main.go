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

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gotest/config"
	"gotest/delivery"
	_ "gotest/docs"
	"gotest/repository"
	"gotest/usecase"
)

func main() {
	// 1. Inisialisasi Go 1.21+ Structured JSON Logger & Load Konfigurasi
	logger := delivery.InitLogger()
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

	// 6. Inisialisasi Gin Engine dengan Custom Structured Logger & Request-ID
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(delivery.RequestIDMiddleware())
	r.Use(delivery.StructuredLoggerMiddleware(logger))
	r.Use(delivery.CORSMiddleware())
	r.Use(delivery.RateLimitMiddleware(5, 10)) // Batas: 5 req/detik, burst 10

	// -------------------------------------------------------------------------
	// SWAGGER OPENAPI DOCUMENTATION ROUTE
	// -------------------------------------------------------------------------
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		accountGroup.POST("/withdraw", ginBankHandler.Withdraw)
		accountGroup.GET("/mutations", ginBankHandler.GetMutations)
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
		fmt.Printf("📖 Interactive Swagger API UI: http://localhost:%s/swagger/index.html\n", cfg.Port)
		fmt.Println("🛡️  Protection: IP Rate Limiter Active (5 req/sec, burst 10)")
		fmt.Println("📊 Observability: Structured JSON Logger (slog) & X-Request-ID Tracking Active")

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
