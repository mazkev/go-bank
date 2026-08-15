package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	DBDriver   string // "sqlite" atau "postgres"
	DBPath     string // Path file untuk SQLite (contoh: bank.db)
	DBHost     string // Host untuk PostgreSQL (contoh: localhost)
	DBPort     string // Port untuk PostgreSQL (contoh: 5432)
	DBUser     string // Username PostgreSQL (contoh: postgres)
	DBPassword string // Password PostgreSQL
	DBName     string // Nama Database PostgreSQL
	DBSSLMode  string // disable, require, dll.
	JWTSecret  string
	GinMode    string
}

func LoadConfig() *Config {
	_ = godotenv.Load() // Memuat variabel dari file .env jika ada

	return &Config{
		Port:       getEnv("PORT", "8080"),
		DBDriver:   getEnv("DB_DRIVER", "sqlite"),
		DBPath:     getEnv("DB_PATH", "bank.db"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "bank_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", "SuperSecretBankingKey2026"),
		GinMode:    getEnv("GIN_MODE", "release"),
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
