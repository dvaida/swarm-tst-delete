package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	defaultModel     = "gemini-embedding-001"
	defaultRateLimit = 60
	defaultBaseURL   = "https://generativelanguage.googleapis.com/v1beta"
	maxRetries       = 3
	initialBackoff   = 1 * time.Second
	backoffMultiplier = 2
)

// GeminiClient is a client for generating embeddings via Gemini API.
type GeminiClient struct {
	apiKey     string
	model      string
	rateLimit  int
	limiter    *rate.Limiter
	httpClient *http.Client
	baseURL    string
}

// Request/Response types for Gemini API
type embedRequest struct {
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}

type embedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

type batchEmbedRequest struct {
	Requests []embedRequest `json:"requests"`
}

type batchEmbedResponse struct {
	Embeddings []struct {
		Values []float32 `json:"values"`
	} `json:"embeddings"`
}

type apiError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// NewGeminiClient creates a new Gemini client for generating embeddings.
func NewGeminiClient(apiKey, model string, rateLimit int) *GeminiClient {
	if model == "" {
		model = defaultModel
	}
	if rateLimit <= 0 {
		rateLimit = defaultRateLimit
	}

	// Convert requests/minute to requests/second for rate.Limiter
	rps := float64(rateLimit) / 60.0
	limiter := rate.NewLimiter(rate.Limit(rps), 1)

	return &GeminiClient{
		apiKey:     apiKey,
		model:      model,
		rateLimit:  rateLimit,
		limiter:    limiter,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    defaultBaseURL,
	}
}

// Embed generates an embedding for a single text.
func (c *GeminiClient) Embed(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, errors.New("text cannot be empty")
	}

	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	req := embedRequest{}
	req.Content.Parts = []struct {
		Text string `json:"text"`
	}{{Text: text}}

	url := fmt.Sprintf("%s/models/%s:embedContent?key=%s", c.baseURL, c.model, c.apiKey)

	var resp embedResponse
	if err := c.doRequestWithRetry(ctx, url, req, &resp); err != nil {
		return nil, err
	}

	return resp.Embedding.Values, nil
}

// EmbedBatch generates embeddings for multiple texts in a single batched request.
func (c *GeminiClient) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, errors.New("texts cannot be empty")
	}

	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	batchReq := batchEmbedRequest{
		Requests: make([]embedRequest, len(texts)),
	}
	for i, text := range texts {
		batchReq.Requests[i].Content.Parts = []struct {
			Text string `json:"text"`
		}{{Text: text}}
	}

	url := fmt.Sprintf("%s/models/%s:batchEmbedContents?key=%s", c.baseURL, c.model, c.apiKey)

	var resp batchEmbedResponse
	if err := c.doRequestWithRetry(ctx, url, batchReq, &resp); err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(resp.Embeddings))
	for i, emb := range resp.Embeddings {
		embeddings[i] = emb.Values
	}

	return embeddings, nil
}

func (c *GeminiClient) doRequestWithRetry(ctx context.Context, url string, body interface{}, result interface{}) error {
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry with exponential backoff
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff *= backoffMultiplier
			}
		}

		err := c.doRequest(ctx, url, body, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		var apiErr *APIError
		if errors.As(err, &apiErr) && !apiErr.IsRetryable() {
			return err
		}

		// Context errors are not retryable
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *GeminiClient) doRequest(ctx context.Context, url string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error.Message != "" {
			return &APIError{
				StatusCode: resp.StatusCode,
				Code:       apiErr.Error.Code,
				Message:    apiErr.Error.Message,
				Status:     apiErr.Error.Status,
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// APIError represents an error from the Gemini API.
type APIError struct {
	StatusCode int
	Code       int
	Message    string
	Status     string
}

func (e *APIError) Error() string {
	if e.Status != "" {
		return fmt.Sprintf("gemini API error (status=%d, code=%d, status=%s): %s", e.StatusCode, e.Code, e.Status, e.Message)
	}
	return fmt.Sprintf("gemini API error (status=%d): %s", e.StatusCode, e.Message)
}

// IsRetryable returns true if the error is transient and the request should be retried.
func (e *APIError) IsRetryable() bool {
	switch e.StatusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	}
	return false
}
