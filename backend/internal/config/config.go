package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	DatabaseURL      string
	AllowedOrigins   string
	SyncBufferSec    int // Network latency buffer before sync start
	CycleDurationSec int // 5-hour cycle duration in seconds (18,000s)
}

// LoadConfig reads configuration from environment variables or .env file
func LoadConfig() *Config {
	// Attempt to load .env file if present
	if err := godotenv.Load(); err != nil {
		// Non-fatal if .env doesn't exist, as env vars might be passed directly
		log.Println("[CONFIG] No .env file found or error loading, using system environment variables")
	}

	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/mediasequencer?sslmode=disable")
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "*")
	
	syncBufferSecStr := getEnv("SYNC_BUFFER_SECONDS", "2")
	syncBufferSec, err := strconv.Atoi(syncBufferSecStr)
	if err != nil || syncBufferSec < 0 {
		syncBufferSec = 2
	}

	cycleDurationSecStr := getEnv("CYCLE_DURATION_SECONDS", "18000") // 5 hours = 18000s
	cycleDurationSec, err := strconv.Atoi(cycleDurationSecStr)
	if err != nil || cycleDurationSec <= 0 {
		cycleDurationSec = 18000
	}

	return &Config{
		Port:             port,
		DatabaseURL:      dbURL,
		AllowedOrigins:   allowedOrigins,
		SyncBufferSec:    syncBufferSec,
		CycleDurationSec: cycleDurationSec,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
