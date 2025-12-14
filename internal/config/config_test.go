package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Set required environment variables
	os.Setenv("TYPESENSE_API_KEY", "test-typesense-key")
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	defer func() {
		os.Unsetenv("TYPESENSE_API_KEY")
		os.Unsetenv("GEMINI_API_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test Typesense defaults
	if cfg.TypesenseURL != "http://localhost:8108" {
		t.Errorf("expected TypesenseURL to be 'http://localhost:8108', got '%s'", cfg.TypesenseURL)
	}
	if cfg.TypesenseCollection != "swarm-index" {
		t.Errorf("expected TypesenseCollection to be 'swarm-index', got '%s'", cfg.TypesenseCollection)
	}

	// Test Gemini defaults
	if cfg.GeminiModel != "gemini-embedding-001" {
		t.Errorf("expected GeminiModel to be 'gemini-embedding-001', got '%s'", cfg.GeminiModel)
	}
	if cfg.GeminiRateLimit != 60 {
		t.Errorf("expected GeminiRateLimit to be 60, got %d", cfg.GeminiRateLimit)
	}

	// Test Worker defaults
	if cfg.Workers != 8 {
		t.Errorf("expected Workers to be 8, got %d", cfg.Workers)
	}
	if cfg.BatchSize != 100 {
		t.Errorf("expected BatchSize to be 100, got %d", cfg.BatchSize)
	}

	// Test SkipFiles default
	expectedSkipFiles := ".env,.setenv,*.pem,*.key,credentials.*"
	if cfg.SkipFiles != expectedSkipFiles {
		t.Errorf("expected SkipFiles to be '%s', got '%s'", expectedSkipFiles, cfg.SkipFiles)
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	// Set all environment variables
	os.Setenv("TYPESENSE_URL", "http://custom:8108")
	os.Setenv("TYPESENSE_API_KEY", "custom-typesense-key")
	os.Setenv("TYPESENSE_COLLECTION", "custom-collection")
	os.Setenv("GEMINI_API_KEY", "custom-gemini-key")
	os.Setenv("GEMINI_MODEL", "custom-model")
	os.Setenv("GEMINI_RATE_LIMIT", "120")
	os.Setenv("SWARM_INDEXER_WORKERS", "16")
	os.Setenv("SWARM_INDEXER_BATCH_SIZE", "200")
	os.Setenv("SWARM_INDEXER_SKIP_FILES", "*.txt,*.log")
	defer func() {
		os.Unsetenv("TYPESENSE_URL")
		os.Unsetenv("TYPESENSE_API_KEY")
		os.Unsetenv("TYPESENSE_COLLECTION")
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("GEMINI_MODEL")
		os.Unsetenv("GEMINI_RATE_LIMIT")
		os.Unsetenv("SWARM_INDEXER_WORKERS")
		os.Unsetenv("SWARM_INDEXER_BATCH_SIZE")
		os.Unsetenv("SWARM_INDEXER_SKIP_FILES")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.TypesenseURL != "http://custom:8108" {
		t.Errorf("expected TypesenseURL to be 'http://custom:8108', got '%s'", cfg.TypesenseURL)
	}
	if cfg.TypesenseAPIKey != "custom-typesense-key" {
		t.Errorf("expected TypesenseAPIKey to be 'custom-typesense-key', got '%s'", cfg.TypesenseAPIKey)
	}
	if cfg.TypesenseCollection != "custom-collection" {
		t.Errorf("expected TypesenseCollection to be 'custom-collection', got '%s'", cfg.TypesenseCollection)
	}
	if cfg.GeminiAPIKey != "custom-gemini-key" {
		t.Errorf("expected GeminiAPIKey to be 'custom-gemini-key', got '%s'", cfg.GeminiAPIKey)
	}
	if cfg.GeminiModel != "custom-model" {
		t.Errorf("expected GeminiModel to be 'custom-model', got '%s'", cfg.GeminiModel)
	}
	if cfg.GeminiRateLimit != 120 {
		t.Errorf("expected GeminiRateLimit to be 120, got %d", cfg.GeminiRateLimit)
	}
	if cfg.Workers != 16 {
		t.Errorf("expected Workers to be 16, got %d", cfg.Workers)
	}
	if cfg.BatchSize != 200 {
		t.Errorf("expected BatchSize to be 200, got %d", cfg.BatchSize)
	}
	if cfg.SkipFiles != "*.txt,*.log" {
		t.Errorf("expected SkipFiles to be '*.txt,*.log', got '%s'", cfg.SkipFiles)
	}
}

func TestLoadConfig_MissingTypesenseAPIKey(t *testing.T) {
	// Unset required variables
	os.Unsetenv("TYPESENSE_API_KEY")
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when TYPESENSE_API_KEY is missing")
	}
}

func TestLoadConfig_MissingGeminiAPIKey(t *testing.T) {
	// Unset required variables
	os.Setenv("TYPESENSE_API_KEY", "test-typesense-key")
	os.Unsetenv("GEMINI_API_KEY")
	defer os.Unsetenv("TYPESENSE_API_KEY")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when GEMINI_API_KEY is missing")
	}
}
