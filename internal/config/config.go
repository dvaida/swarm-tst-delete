package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the indexer
type Config struct {
	TypesenseURL        string
	TypesenseAPIKey     string
	TypesenseCollection string
	GeminiAPIKey        string
	GeminiModel         string
	GeminiRateLimit     int
	Workers             int
	BatchSize           int
	SkipFiles           []string
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		TypesenseURL:        getEnv("TYPESENSE_URL", "http://localhost:8108"),
		TypesenseAPIKey:     os.Getenv("TYPESENSE_API_KEY"),
		TypesenseCollection: getEnv("TYPESENSE_COLLECTION", "swarm-index"),
		GeminiAPIKey:        os.Getenv("GEMINI_API_KEY"),
		GeminiModel:         getEnv("GEMINI_MODEL", "gemini-embedding-001"),
		GeminiRateLimit:     getEnvInt("GEMINI_RATE_LIMIT", 60),
		Workers:             getEnvInt("SWARM_INDEXER_WORKERS", 8),
		BatchSize:           getEnvInt("SWARM_INDEXER_BATCH_SIZE", 100),
	}
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
