package indexer

import (
	"context"
)

// IndexedChunk represents a document in Typesense
type IndexedChunk struct {
	ID          string    `json:"id"`
	FilePath    string    `json:"file_path"`
	ProjectPath string    `json:"project_path"`
	ProjectType string    `json:"project_type"`
	Language    string    `json:"language"`
	ChunkType   string    `json:"chunk_type"`
	Content     string    `json:"content"`
	Embedding   []float32 `json:"embedding"`
	StartLine   int       `json:"start_line"`
	EndLine     int       `json:"end_line"`
	LastIndexed int64     `json:"last_indexed"`
}

// TypesenseClient wraps the Typesense client
type TypesenseClient struct {
	url        string
	apiKey     string
	collection string
}

// NewTypesenseClient creates a new TypesenseClient
func NewTypesenseClient(url, apiKey, collection string) *TypesenseClient {
	return &TypesenseClient{
		url:        url,
		apiKey:     apiKey,
		collection: collection,
	}
}

// UpsertBatch upserts a batch of chunks to Typesense
func (t *TypesenseClient) UpsertBatch(ctx context.Context, chunks []IndexedChunk) error {
	// Stub implementation
	return nil
}

// Delete deletes a document from Typesense
func (t *TypesenseClient) Delete(ctx context.Context, id string) error {
	// Stub implementation
	return nil
}

// EnsureCollection ensures the collection exists with correct schema
func (t *TypesenseClient) EnsureCollection(ctx context.Context) error {
	// Stub implementation
	return nil
}
