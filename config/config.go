package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	JWTSecret string
	DBPath    string
	GinMode   string
}

// LoadConfig membaca environment variables dari file .env
func LoadConfig() *Config {
	// Load file .env jika ada (tidak error jika file tidak ditemukan di prod)
	_ = godotenv.Load()

	return &Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", "DefaultFallbackSecretKey"),
		DBPath:    getEnv("DB_PATH", "bank.db"),
		GinMode:   getEnv("GIN_MODE", "debug"),
	}
}

// Helper untuk membaca env var dengan nilai default (fallback)
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
