package config

import "os"

// Config holds all configuration settings
type Config struct {
	TypesenseURL        string
	TypesenseAPIKey     string
	TypesenseCollection string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		TypesenseURL:        getEnv("TYPESENSE_URL", "http://localhost:8108"),
		TypesenseAPIKey:     getEnv("TYPESENSE_API_KEY", ""),
		TypesenseCollection: getEnv("TYPESENSE_COLLECTION", "swarm-index"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
