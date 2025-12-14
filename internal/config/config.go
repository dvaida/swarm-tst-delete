package config

import (
	"errors"
	"os"
	"strconv"
)

// Config holds all configuration for swarm-indexer
type Config struct {
	// Typesense settings
	TypesenseURL        string
	TypesenseAPIKey     string
	TypesenseCollection string

	// Gemini settings
	GeminiAPIKey    string
	GeminiModel     string
	GeminiRateLimit int

	// Worker settings
	Workers   int
	BatchSize int

	// Skip files pattern
	SkipFiles string
}

// Load loads configuration from environment variables
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
		SkipFiles:           getEnv("SWARM_INDEXER_SKIP_FILES", ".env,.setenv,*.pem,*.key,credentials.*"),
	}

	if cfg.TypesenseAPIKey == "" {
		return nil, errors.New("TYPESENSE_API_KEY is required")
	}
	if cfg.GeminiAPIKey == "" {
		return nil, errors.New("GEMINI_API_KEY is required")
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
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
