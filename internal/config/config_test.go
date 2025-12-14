package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear relevant ENV vars to test defaults
	clearEnvVars()
	// Set required vars to avoid validation errors
	os.Setenv("TYPESENSE_API_KEY", "test-key")
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer clearEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Test default values
	if cfg.TypesenseURL != "http://localhost:8108" {
		t.Errorf("TypesenseURL = %q, want %q", cfg.TypesenseURL, "http://localhost:8108")
	}
	if cfg.TypesenseCollection != "swarm-index" {
		t.Errorf("TypesenseCollection = %q, want %q", cfg.TypesenseCollection, "swarm-index")
	}
	if cfg.GeminiModel != "gemini-embedding-001" {
		t.Errorf("GeminiModel = %q, want %q", cfg.GeminiModel, "gemini-embedding-001")
	}
	if cfg.GeminiRateLimit != 60 {
		t.Errorf("GeminiRateLimit = %d, want %d", cfg.GeminiRateLimit, 60)
	}
	if cfg.Workers != 8 {
		t.Errorf("Workers = %d, want %d", cfg.Workers, 8)
	}
	if cfg.BatchSize != 100 {
		t.Errorf("BatchSize = %d, want %d", cfg.BatchSize, 100)
	}
}

func TestLoad_FromEnvVars(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	// Set all ENV vars
	os.Setenv("TYPESENSE_URL", "http://custom:9999")
	os.Setenv("TYPESENSE_API_KEY", "my-typesense-key")
	os.Setenv("TYPESENSE_COLLECTION", "custom-collection")
	os.Setenv("GEMINI_API_KEY", "my-gemini-key")
	os.Setenv("GEMINI_MODEL", "custom-model")
	os.Setenv("GEMINI_RATE_LIMIT", "120")
	os.Setenv("SWARM_INDEXER_WORKERS", "16")
	os.Setenv("SWARM_INDEXER_BATCH_SIZE", "200")
	os.Setenv("SWARM_INDEXER_SKIP_FILES", ".secret,*.private")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.TypesenseURL != "http://custom:9999" {
		t.Errorf("TypesenseURL = %q, want %q", cfg.TypesenseURL, "http://custom:9999")
	}
	if cfg.TypesenseAPIKey != "my-typesense-key" {
		t.Errorf("TypesenseAPIKey = %q, want %q", cfg.TypesenseAPIKey, "my-typesense-key")
	}
	if cfg.TypesenseCollection != "custom-collection" {
		t.Errorf("TypesenseCollection = %q, want %q", cfg.TypesenseCollection, "custom-collection")
	}
	if cfg.GeminiAPIKey != "my-gemini-key" {
		t.Errorf("GeminiAPIKey = %q, want %q", cfg.GeminiAPIKey, "my-gemini-key")
	}
	if cfg.GeminiModel != "custom-model" {
		t.Errorf("GeminiModel = %q, want %q", cfg.GeminiModel, "custom-model")
	}
	if cfg.GeminiRateLimit != 120 {
		t.Errorf("GeminiRateLimit = %d, want %d", cfg.GeminiRateLimit, 120)
	}
	if cfg.Workers != 16 {
		t.Errorf("Workers = %d, want %d", cfg.Workers, 16)
	}
	if cfg.BatchSize != 200 {
		t.Errorf("BatchSize = %d, want %d", cfg.BatchSize, 200)
	}
}

func TestLoad_RequiredTypesenseAPIKey(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	os.Setenv("GEMINI_API_KEY", "test-key")
	// TYPESENSE_API_KEY not set

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error when TYPESENSE_API_KEY is missing")
	}
}

func TestLoad_RequiredGeminiAPIKey(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	os.Setenv("TYPESENSE_API_KEY", "test-key")
	// GEMINI_API_KEY not set

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error when GEMINI_API_KEY is missing")
	}
}

func TestLoad_SkipFilesParsesCsv(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	os.Setenv("TYPESENSE_API_KEY", "test-key")
	os.Setenv("GEMINI_API_KEY", "test-key")
	os.Setenv("SWARM_INDEXER_SKIP_FILES", ".env,.setenv,*.pem,*.key,credentials.*")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	expected := []string{".env", ".setenv", "*.pem", "*.key", "credentials.*"}
	if len(cfg.SkipFiles) != len(expected) {
		t.Fatalf("SkipFiles length = %d, want %d", len(cfg.SkipFiles), len(expected))
	}
	for i, pattern := range expected {
		if cfg.SkipFiles[i] != pattern {
			t.Errorf("SkipFiles[%d] = %q, want %q", i, cfg.SkipFiles[i], pattern)
		}
	}
}

func TestLoad_DefaultSkipFiles(t *testing.T) {
	clearEnvVars()
	defer clearEnvVars()

	os.Setenv("TYPESENSE_API_KEY", "test-key")
	os.Setenv("GEMINI_API_KEY", "test-key")
	// SWARM_INDEXER_SKIP_FILES not set

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	expected := []string{".env", ".setenv", "*.pem", "*.key", "credentials.*"}
	if len(cfg.SkipFiles) != len(expected) {
		t.Fatalf("SkipFiles length = %d, want %d", len(cfg.SkipFiles), len(expected))
	}
}

func clearEnvVars() {
	os.Unsetenv("TYPESENSE_URL")
	os.Unsetenv("TYPESENSE_API_KEY")
	os.Unsetenv("TYPESENSE_COLLECTION")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("GEMINI_MODEL")
	os.Unsetenv("GEMINI_RATE_LIMIT")
	os.Unsetenv("SWARM_INDEXER_WORKERS")
	os.Unsetenv("SWARM_INDEXER_BATCH_SIZE")
	os.Unsetenv("SWARM_INDEXER_SKIP_FILES")
}
