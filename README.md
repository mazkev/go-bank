# 🏦 Enterprise Go Banking System (Clean Architecture)

[![Go CI](https://github.com/mazkev/go-bank/actions/workflows/ci.yml/badge.svg)](https://github.com/mazkev/go-bank/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Swagger UI](https://img.shields.io/badge/Swagger-OpenAPI%203.0-green.svg)](http://localhost:8080/swagger/index.html)

Sistem backend perbankan skala produksi (*Enterprise Core Banking System*) yang dibangun menggunakan **Go**, **Clean Architecture (Uncle Bob Standard)**, **Gin Web Framework**, dan **GORM SQLite Database**.

---

## 🏛️ Arsitektur Sistem (Clean Architecture 4-Layer)

Aplikasi ini memisahkan tanggung jawab kode ke dalam 4 layer terisolasi sesuai prinsip *Dependency Inversion*:

```
├── domain/         # Entities, Interface Contracts, Domain Errors
├── repository/     # Data Access Layer (GORM SQLite & In-Memory Storage)
├── usecase/        # Business Logic Layer (ACID Transactions, Hashing, JWT)
├── delivery/       # Presentation & HTTP Layer (Gin Handlers & Middlewares)
├── config/         # Environment Variables Loader (.env)
└── docs/           # OpenAPI / Swagger Specification
```

---

## ⚡ Fitur-Fitur Utama (Production-Grade Features)

1. **Clean Architecture & Dependency Injection:** Seluruh komponen terhubung melalui abstraksi antarmuka (*interface*).
2. **Atomic ACID Database Transaction & Rollback (`ExecTx`):** Transaksi transfer saldo dan mutasi dibungkus dalam 1 transaksi atomik dengan *automatic rollback* jika terjadi kegagalan.
3. **Keamanan Otentikasi & Otorisasi:**
   - Enkripsi password menggunakan **Bcrypt** (`golang.org/x/crypto/bcrypt`).
   - Autentikasi stateless menggunakan **JWT Bearer Token** (`golang-jwt/jwt/v5`).
   - **PIN Transaksi 6-Digit** (Bcrypt-hashed) untuk verifikasi transfer dan penarikan tunai (*withdraw*).
4. **Proteksi Anti-Spam & DDoS:** **IP Rate Limiter Middleware** berbasis algoritma *Token Bucket* (`golang.org/x/time/rate`).
5. **Observability & Audit Trail:**
   - **Request-ID Correlation Middleware:** Header `X-Request-ID` (UUID) di setiap request HTTP.
   - **Structured JSON Logging:** Menggunakan `log/slog` bawaan Go 1.21+.
6. **Dokumentasi API Interaktif:** **Swagger / OpenAPI 3.0** (`swaggo/swag`) di `http://localhost:8080/swagger/index.html`.
7. **Production Graceful Shutdown:** Menangani sinyal `SIGINT` / `SIGTERM` dengan `context.WithTimeout(5s)`.
8. **CI/CD Automation:** **GitHub Actions Pipeline** untuk pengujian otomatis (*Unit & Integration Tests*) serta *build validation*.

---

## 🚀 Cara Menjalankan Aplikasi

### 1. Prasyarat
- Go `v1.21` atau lebih baru terpasang di sistem Anda.

### 2. Jalankan Server Lokal
```bash
# Clone repository
git clone https://github.com/mazkev/go-bank.git
cd go-bank

# Jalankan aplikasi
go run .
```
Server akan berjalan di `http://localhost:8080`.

### 3. Akses Swagger Documentation
Buka browser Anda dan kunjungi:
👉 **[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

---

## 🧪 Menjalankan Pengujian (Testing)

```bash
# Menjalankan seluruh Unit & Integration Test
go test -v -race ./...

# Menjalankan Usecase Mocking Unit Tests
go test -v ./usecase

# Menjalankan HTTP Integration Tests
go test -v ./delivery
```

---

## 📄 REST Client API Test Suite

File **[`test.http`](test.http)** telah disediakan untuk pengujian cepat langsung dari VS Code menggunakan ekstensi *REST Client*:
- Registrasi Nasabah Baru (Budi & Siti)
- Login & Pengambilan JWT Bearer Token
- Cek Detail Rekening
- Transfer Saldo dengan Verifikasi PIN 6-digit
- Tarik Tunai (*Withdraw*) dengan Verifikasi PIN 6-digit
- Cek Mutasi & Histori Transaksi Rekening

---

## 📜 Lisensi
Didistribusikan di bawah lisensi MIT.
