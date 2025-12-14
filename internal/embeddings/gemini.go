package embeddings

import (
	"context"
)

// GeminiClient generates embeddings using Gemini API
type GeminiClient struct {
	apiKey    string
	model     string
	rateLimit int
}

// New creates a new GeminiClient
func New(apiKey, model string, rateLimit int) *GeminiClient {
	return &GeminiClient{
		apiKey:    apiKey,
		model:     model,
		rateLimit: rateLimit,
	}
}

// GenerateEmbedding generates an embedding for a single text
func (g *GeminiClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Stub: return a fixed-size embedding
	return make([]float32, 768), nil
}

// GenerateEmbeddings generates embeddings for multiple texts (batched)
func (g *GeminiClient) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		embeddings[i] = make([]float32, 768)
	}
	return embeddings, nil
}
