package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for swarm-indexer
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

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		TypesenseURL:        getEnvOrDefault("TYPESENSE_URL", "http://localhost:8108"),
		TypesenseAPIKey:     os.Getenv("TYPESENSE_API_KEY"),
		TypesenseCollection: getEnvOrDefault("TYPESENSE_COLLECTION", "swarm-index"),
		GeminiAPIKey:        os.Getenv("GEMINI_API_KEY"),
		GeminiModel:         getEnvOrDefault("GEMINI_MODEL", "gemini-embedding-001"),
		GeminiRateLimit:     getEnvIntOrDefault("GEMINI_RATE_LIMIT", 60),
		Workers:             getEnvIntOrDefault("SWARM_INDEXER_WORKERS", 8),
		BatchSize:           getEnvIntOrDefault("SWARM_INDEXER_BATCH_SIZE", 100),
		SkipFiles:           getEnvSliceOrDefault("SWARM_INDEXER_SKIP_FILES", []string{".env", ".setenv", "*.pem", "*.key", "credentials.*"}),
	}

	if cfg.TypesenseAPIKey == "" {
		return nil, errors.New("TYPESENSE_API_KEY is required")
	}
	if cfg.GeminiAPIKey == "" {
		return nil, errors.New("GEMINI_API_KEY is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvSliceOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
