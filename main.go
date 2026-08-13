package main

import (
	"context"
	"fmt"
	"net/http"

	"gotest/delivery"
	"gotest/repository"
	"gotest/usecase"
)

func main() {
	// 1. Inisialisasi Repository Layer (Data Access)
	bankRepo := repository.NewMemoryBankRepository()

	// 2. Inisialisasi Usecase Layer (Business Logic) & Inject Repository
	bankUsecase := usecase.NewBankUsecase(bankRepo)

	// 3. Pre-populate Data Awal menggunakan Usecase
	ctx := context.Background()
	acc1, _ := bankUsecase.CreateAccount(ctx, "Budi Santoso", 1000000)
	acc2, _ := bankUsecase.CreateAccount(ctx, "Siti Aminah", 500000)

	fmt.Printf(" Data Awal (Clean Arch) Berhasil Dibuat:\n")
	fmt.Printf("   1. Akun %s (%s) - Saldo: Rp %.0f\n", acc1.ID, acc1.OwnerName, acc1.Balance)
	fmt.Printf("   2. Akun %s (%s) - Saldo: Rp %.0f\n", acc2.ID, acc2.OwnerName, acc2.Balance)

	// 4. Inisialisasi HTTP Handler Layer & Inject Usecase
	bankHandler := delivery.NewBankHandler(bankUsecase)

	// 5. Registrasi Route HTTP Endpoint
	http.HandleFunc("/accounts/create", bankHandler.CreateAccount)
	http.HandleFunc("/accounts/get", bankHandler.GetAccountByID)
	http.HandleFunc("/accounts/transfer", bankHandler.Transfer)

	port := ":8080"
	fmt.Printf("\n🚀 Clean Architecture Bank API berjalan di http://localhost%s\n", port)
	fmt.Println("Tekan Ctrl+C untuk menghentikan server.")

	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
