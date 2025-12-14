# Plan: Issue #24 - Implement Gemini embeddings with rate limiting

## Overview
Implement a Gemini API client for generating text embeddings with proper rate limiting and exponential backoff retry logic.

## Files to Create
- `internal/embeddings/gemini.go` - Main implementation
- `internal/embeddings/gemini_test.go` - Integration tests
- `go.mod` - Go module file (project doesn't have one yet)

## Integration Tests to Write

1. **TestNewGeminiClient** - Verifies client creation with valid config
2. **TestNewGeminiClient_InvalidAPIKey** - Verifies error on empty API key
3. **TestEmbed_Success** - Tests single text embedding generation (uses mock server)
4. **TestEmbed_EmptyText** - Tests error handling for empty input
5. **TestEmbedBatch_Success** - Tests batch embedding generation
6. **TestEmbedBatch_EmptyBatch** - Tests error handling for empty batch
7. **TestEmbed_RateLimiting** - Verifies rate limiter respects configured limit
8. **TestEmbed_RetryOn429** - Verifies retry with backoff on 429 responses
9. **TestEmbed_RetryOn503** - Verifies retry with backoff on 503 responses
10. **TestEmbed_APIError** - Tests graceful handling of API errors with clear messages

## Implementation Approach

### Data Structures
```go
type GeminiClient struct {
    apiKey      string
    model       string
    rateLimit   int
    limiter     *rate.Limiter
    httpClient  *http.Client
    baseURL     string
}
```

### Key Implementation Details

1. **Rate Limiting**: Use `golang.org/x/time/rate` token bucket limiter
   - Convert requests/minute to requests/second for rate.Limiter
   - Call `limiter.Wait(ctx)` before each API request

2. **Retry Logic**: Exponential backoff for transient errors
   - Initial delay: 1 second
   - Max retries: 3
   - Backoff multiplier: 2
   - Retryable status codes: 429, 500, 502, 503, 504

3. **API Integration**:
   - Use Gemini REST API endpoint for embeddings
   - Endpoint: `POST /v1beta/models/{model}:embedContent`
   - Request body includes text content
   - Response contains embedding vector

4. **Batch Processing**:
   - Gemini supports `batchEmbedContents` endpoint
   - Process multiple texts in a single API call
   - Endpoint: `POST /v1beta/models/{model}:batchEmbedContents`

## Test Strategy

Tests will use `httptest.Server` to mock the Gemini API responses, allowing us to:
- Test success cases with predetermined embedding vectors
- Simulate 429/503 errors to test retry logic
- Verify rate limiting by checking request timing
- Test error handling without hitting real API

## Configuration
- `GEMINI_API_KEY` - Required, no default
- `GEMINI_MODEL` - Default: "gemini-embedding-001"
- `GEMINI_RATE_LIMIT` - Default: 60 requests/min
