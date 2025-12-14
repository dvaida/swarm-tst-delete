package config

import (
	"os"
	"testing"
)

func TestLoad_WithAllEnvVarsSet(t *testing.T) {
	// Set all environment variables
	os.Setenv("TYPESENSE_URL", "http://custom:8108")
	os.Setenv("TYPESENSE_API_KEY", "test-typesense-key")
	os.Setenv("TYPESENSE_COLLECTION", "custom-collection")
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")
	os.Setenv("GEMINI_MODEL", "custom-model")
	os.Setenv("GEMINI_RATE_LIMIT", "120")
	os.Setenv("SWARM_INDEXER_WORKERS", "16")
	os.Setenv("SWARM_INDEXER_BATCH_SIZE", "200")
	os.Setenv("SWARM_INDEXER_SKIP_FILES", ".secret,private.key")
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
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.TypesenseURL != "http://custom:8108" {
		t.Errorf("TypesenseURL = %q, want %q", cfg.TypesenseURL, "http://custom:8108")
	}
	if cfg.TypesenseAPIKey != "test-typesense-key" {
		t.Errorf("TypesenseAPIKey = %q, want %q", cfg.TypesenseAPIKey, "test-typesense-key")
	}
	if cfg.TypesenseCollection != "custom-collection" {
		t.Errorf("TypesenseCollection = %q, want %q", cfg.TypesenseCollection, "custom-collection")
	}
	if cfg.GeminiAPIKey != "test-gemini-key" {
		t.Errorf("GeminiAPIKey = %q, want %q", cfg.GeminiAPIKey, "test-gemini-key")
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
	if len(cfg.SkipFiles) != 2 || cfg.SkipFiles[0] != ".secret" || cfg.SkipFiles[1] != "private.key" {
		t.Errorf("SkipFiles = %v, want %v", cfg.SkipFiles, []string{".secret", "private.key"})
	}
}

func TestLoad_WithDefaults(t *testing.T) {
	// Only set required variables
	os.Setenv("TYPESENSE_API_KEY", "required-key")
	os.Setenv("GEMINI_API_KEY", "required-key")
	// Unset optional variables to ensure defaults are used
	os.Unsetenv("TYPESENSE_URL")
	os.Unsetenv("TYPESENSE_COLLECTION")
	os.Unsetenv("GEMINI_MODEL")
	os.Unsetenv("GEMINI_RATE_LIMIT")
	os.Unsetenv("SWARM_INDEXER_WORKERS")
	os.Unsetenv("SWARM_INDEXER_BATCH_SIZE")
	os.Unsetenv("SWARM_INDEXER_SKIP_FILES")
	defer func() {
		os.Unsetenv("TYPESENSE_API_KEY")
		os.Unsetenv("GEMINI_API_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.TypesenseURL != "http://localhost:8108" {
		t.Errorf("TypesenseURL = %q, want default %q", cfg.TypesenseURL, "http://localhost:8108")
	}
	if cfg.TypesenseCollection != "swarm-index" {
		t.Errorf("TypesenseCollection = %q, want default %q", cfg.TypesenseCollection, "swarm-index")
	}
	if cfg.GeminiModel != "gemini-embedding-001" {
		t.Errorf("GeminiModel = %q, want default %q", cfg.GeminiModel, "gemini-embedding-001")
	}
	if cfg.GeminiRateLimit != 60 {
		t.Errorf("GeminiRateLimit = %d, want default %d", cfg.GeminiRateLimit, 60)
	}
	if cfg.Workers != 8 {
		t.Errorf("Workers = %d, want default %d", cfg.Workers, 8)
	}
	if cfg.BatchSize != 100 {
		t.Errorf("BatchSize = %d, want default %d", cfg.BatchSize, 100)
	}
	expectedSkipFiles := []string{".env", ".setenv", "*.pem", "*.key", "credentials.*"}
	if len(cfg.SkipFiles) != len(expectedSkipFiles) {
		t.Errorf("SkipFiles = %v, want default %v", cfg.SkipFiles, expectedSkipFiles)
	}
}

func TestLoad_MissingTypesenseAPIKey(t *testing.T) {
	os.Unsetenv("TYPESENSE_API_KEY")
	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should return error when TYPESENSE_API_KEY is missing")
	}
}

func TestLoad_MissingGeminiAPIKey(t *testing.T) {
	os.Setenv("TYPESENSE_API_KEY", "test-key")
	os.Unsetenv("GEMINI_API_KEY")
	defer os.Unsetenv("TYPESENSE_API_KEY")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should return error when GEMINI_API_KEY is missing")
	}
}

func TestLoad_SkipFilesCommaSeparated(t *testing.T) {
	os.Setenv("TYPESENSE_API_KEY", "test-key")
	os.Setenv("GEMINI_API_KEY", "test-key")
	os.Setenv("SWARM_INDEXER_SKIP_FILES", "a.txt,b.txt,c.txt")
	defer func() {
		os.Unsetenv("TYPESENSE_API_KEY")
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("SWARM_INDEXER_SKIP_FILES")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	expected := []string{"a.txt", "b.txt", "c.txt"}
	if len(cfg.SkipFiles) != 3 {
		t.Fatalf("SkipFiles length = %d, want %d", len(cfg.SkipFiles), 3)
	}
	for i, v := range expected {
		if cfg.SkipFiles[i] != v {
			t.Errorf("SkipFiles[%d] = %q, want %q", i, cfg.SkipFiles[i], v)
		}
	}
}

func TestLoad_IntegerParsing(t *testing.T) {
	os.Setenv("TYPESENSE_API_KEY", "test-key")
	os.Setenv("GEMINI_API_KEY", "test-key")
	os.Setenv("GEMINI_RATE_LIMIT", "30")
	os.Setenv("SWARM_INDEXER_WORKERS", "4")
	os.Setenv("SWARM_INDEXER_BATCH_SIZE", "50")
	defer func() {
		os.Unsetenv("TYPESENSE_API_KEY")
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("GEMINI_RATE_LIMIT")
		os.Unsetenv("SWARM_INDEXER_WORKERS")
		os.Unsetenv("SWARM_INDEXER_BATCH_SIZE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.GeminiRateLimit != 30 {
		t.Errorf("GeminiRateLimit = %d, want %d", cfg.GeminiRateLimit, 30)
	}
	if cfg.Workers != 4 {
		t.Errorf("Workers = %d, want %d", cfg.Workers, 4)
	}
	if cfg.BatchSize != 50 {
		t.Errorf("BatchSize = %d, want %d", cfg.BatchSize, 50)
	}
}
