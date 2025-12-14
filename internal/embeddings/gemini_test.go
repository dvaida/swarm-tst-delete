package embeddings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// Mock response structures matching Gemini API
type mockEmbeddingResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

type mockBatchEmbeddingResponse struct {
	Embeddings []struct {
		Values []float32 `json:"values"`
	} `json:"embeddings"`
}

type mockErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

func TestNewGeminiClient(t *testing.T) {
	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 60)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.apiKey != "test-api-key" {
		t.Errorf("expected apiKey 'test-api-key', got '%s'", client.apiKey)
	}
	if client.model != "gemini-embedding-001" {
		t.Errorf("expected model 'gemini-embedding-001', got '%s'", client.model)
	}
	if client.rateLimit != 60 {
		t.Errorf("expected rateLimit 60, got %d", client.rateLimit)
	}
}

func TestNewGeminiClient_DefaultModel(t *testing.T) {
	client := NewGeminiClient("test-api-key", "", 60)

	if client.model != "gemini-embedding-001" {
		t.Errorf("expected default model 'gemini-embedding-001', got '%s'", client.model)
	}
}

func TestNewGeminiClient_DefaultRateLimit(t *testing.T) {
	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 0)

	if client.rateLimit != 60 {
		t.Errorf("expected default rateLimit 60, got %d", client.rateLimit)
	}
}

func TestEmbed_Success(t *testing.T) {
	expectedEmbedding := []float32{0.1, 0.2, 0.3, 0.4, 0.5}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		resp := mockEmbeddingResponse{}
		resp.Embedding.Values = expectedEmbedding
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 60)
	client.baseURL = server.URL

	ctx := context.Background()
	embedding, err := client.Embed(ctx, "test text")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(embedding) != len(expectedEmbedding) {
		t.Fatalf("expected %d values, got %d", len(expectedEmbedding), len(embedding))
	}
	for i, v := range embedding {
		if v != expectedEmbedding[i] {
			t.Errorf("embedding[%d]: expected %f, got %f", i, expectedEmbedding[i], v)
		}
	}
}

func TestEmbed_EmptyText(t *testing.T) {
	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 60)

	ctx := context.Background()
	_, err := client.Embed(ctx, "")

	if err == nil {
		t.Fatal("expected error for empty text")
	}
	if err.Error() != "text cannot be empty" {
		t.Errorf("expected 'text cannot be empty' error, got '%s'", err.Error())
	}
}

func TestEmbed_InvalidAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := mockErrorResponse{}
		resp.Error.Code = 401
		resp.Error.Message = "API key not valid"
		resp.Error.Status = "UNAUTHENTICATED"
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("invalid-key", "gemini-embedding-001", 60)
	client.baseURL = server.URL

	ctx := context.Background()
	_, err := client.Embed(ctx, "test text")

	if err == nil {
		t.Fatal("expected error for invalid API key")
	}
}

func TestEmbedBatch_Success(t *testing.T) {
	texts := []string{"text one", "text two", "text three"}
	expectedEmbeddings := [][]float32{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
		{0.7, 0.8, 0.9},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := mockBatchEmbeddingResponse{
			Embeddings: make([]struct {
				Values []float32 `json:"values"`
			}, len(expectedEmbeddings)),
		}
		for i, emb := range expectedEmbeddings {
			resp.Embeddings[i].Values = emb
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 60)
	client.baseURL = server.URL

	ctx := context.Background()
	embeddings, err := client.EmbedBatch(ctx, texts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(embeddings) != len(texts) {
		t.Fatalf("expected %d embeddings, got %d", len(texts), len(embeddings))
	}
	for i, emb := range embeddings {
		if len(emb) != len(expectedEmbeddings[i]) {
			t.Errorf("embedding[%d]: expected %d values, got %d", i, len(expectedEmbeddings[i]), len(emb))
		}
		for j, v := range emb {
			if v != expectedEmbeddings[i][j] {
				t.Errorf("embedding[%d][%d]: expected %f, got %f", i, j, expectedEmbeddings[i][j], v)
			}
		}
	}
}

func TestEmbedBatch_EmptyBatch(t *testing.T) {
	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 60)

	ctx := context.Background()
	_, err := client.EmbedBatch(ctx, []string{})

	if err == nil {
		t.Fatal("expected error for empty batch")
	}
	if err.Error() != "texts cannot be empty" {
		t.Errorf("expected 'texts cannot be empty' error, got '%s'", err.Error())
	}
}

func TestEmbedBatch_NilBatch(t *testing.T) {
	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 60)

	ctx := context.Background()
	_, err := client.EmbedBatch(ctx, nil)

	if err == nil {
		t.Fatal("expected error for nil batch")
	}
}

func TestEmbed_RateLimiting(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		resp := mockEmbeddingResponse{}
		resp.Embedding.Values = []float32{0.1, 0.2, 0.3}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 120 requests/minute = 2 requests/second
	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 120)
	client.baseURL = server.URL

	ctx := context.Background()

	// Make 3 requests and measure time
	start := time.Now()
	for i := 0; i < 3; i++ {
		_, err := client.Embed(ctx, "test text")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	elapsed := time.Since(start)

	// With 2 req/sec rate limit, 3 requests should take at least 1 second
	// (first request immediate, then wait 0.5s, then wait 0.5s)
	if elapsed < 900*time.Millisecond {
		t.Errorf("rate limiting not working: 3 requests completed in %v (expected ~1s)", elapsed)
	}

	if atomic.LoadInt32(&requestCount) != 3 {
		t.Errorf("expected 3 requests, got %d", requestCount)
	}
}

func TestEmbed_RetryOn429(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count <= 2 {
			// First two requests return 429
			w.WriteHeader(http.StatusTooManyRequests)
			resp := mockErrorResponse{}
			resp.Error.Code = 429
			resp.Error.Message = "Resource exhausted"
			resp.Error.Status = "RESOURCE_EXHAUSTED"
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// Third request succeeds
		resp := mockEmbeddingResponse{}
		resp.Embedding.Values = []float32{0.1, 0.2, 0.3}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 6000) // High rate limit to avoid rate limit delays
	client.baseURL = server.URL

	ctx := context.Background()
	embedding, err := client.Embed(ctx, "test text")

	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if len(embedding) != 3 {
		t.Errorf("expected 3 values, got %d", len(embedding))
	}
	if atomic.LoadInt32(&requestCount) != 3 {
		t.Errorf("expected 3 requests (2 retries), got %d", requestCount)
	}
}

func TestEmbed_RetryOn503(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count == 1 {
			// First request returns 503
			w.WriteHeader(http.StatusServiceUnavailable)
			resp := mockErrorResponse{}
			resp.Error.Code = 503
			resp.Error.Message = "Service unavailable"
			resp.Error.Status = "UNAVAILABLE"
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// Second request succeeds
		resp := mockEmbeddingResponse{}
		resp.Embedding.Values = []float32{0.1, 0.2, 0.3}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 6000)
	client.baseURL = server.URL

	ctx := context.Background()
	embedding, err := client.Embed(ctx, "test text")

	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if len(embedding) != 3 {
		t.Errorf("expected 3 values, got %d", len(embedding))
	}
	if atomic.LoadInt32(&requestCount) != 2 {
		t.Errorf("expected 2 requests (1 retry), got %d", requestCount)
	}
}

func TestEmbed_MaxRetriesExceeded(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		// Always return 503
		w.WriteHeader(http.StatusServiceUnavailable)
		resp := mockErrorResponse{}
		resp.Error.Code = 503
		resp.Error.Message = "Service unavailable"
		resp.Error.Status = "UNAVAILABLE"
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 6000)
	client.baseURL = server.URL

	ctx := context.Background()
	_, err := client.Embed(ctx, "test text")

	if err == nil {
		t.Fatal("expected error after max retries exceeded")
	}
	// Should have made 4 requests: initial + 3 retries
	if atomic.LoadInt32(&requestCount) != 4 {
		t.Errorf("expected 4 requests (initial + 3 retries), got %d", requestCount)
	}
}

func TestEmbed_NonRetryableError(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		// Return 400 Bad Request - not retryable
		w.WriteHeader(http.StatusBadRequest)
		resp := mockErrorResponse{}
		resp.Error.Code = 400
		resp.Error.Message = "Invalid request"
		resp.Error.Status = "INVALID_ARGUMENT"
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 6000)
	client.baseURL = server.URL

	ctx := context.Background()
	_, err := client.Embed(ctx, "test text")

	if err == nil {
		t.Fatal("expected error for bad request")
	}
	// Should only make 1 request - no retries for 400
	if atomic.LoadInt32(&requestCount) != 1 {
		t.Errorf("expected 1 request (no retries for 400), got %d", requestCount)
	}
}

func TestEmbed_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Simulate slow response
		resp := mockEmbeddingResponse{}
		resp.Embedding.Values = []float32{0.1, 0.2, 0.3}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("test-api-key", "gemini-embedding-001", 6000)
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.Embed(ctx, "test text")

	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestEmbed_APIKeyInHeader(t *testing.T) {
	var receivedAPIKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAPIKey = r.URL.Query().Get("key")
		resp := mockEmbeddingResponse{}
		resp.Embedding.Values = []float32{0.1, 0.2, 0.3}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewGeminiClient("my-secret-key", "gemini-embedding-001", 60)
	client.baseURL = server.URL

	ctx := context.Background()
	_, err := client.Embed(ctx, "test text")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedAPIKey != "my-secret-key" {
		t.Errorf("expected API key 'my-secret-key' in query param, got '%s'", receivedAPIKey)
	}
}
