package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const defaultBatchSize = 100

// IndexedChunk represents a chunk of code or text indexed in Typesense.
type IndexedChunk struct {
	ID          string    `json:"id"`           // hash of path+offset
	FilePath    string    `json:"file_path"`
	ProjectPath string    `json:"project_path"`
	ProjectType string    `json:"project_type"` // go, node, python, etc.
	Language    string    `json:"language"`
	ChunkType   string    `json:"chunk_type"` // function, class, paragraph
	Content     string    `json:"content"`
	Embedding   []float32 `json:"embedding"` // Gemini vector
	StartLine   int       `json:"start_line"`
	EndLine     int       `json:"end_line"`
	LastIndexed int64     `json:"last_indexed"` // unix timestamp
}

// TypesenseClient wraps the Typesense client for indexing and searching.
type TypesenseClient struct {
	url        string
	apiKey     string
	collection string
	batchSize  int
	httpClient *http.Client
}

// NewTypesenseClient creates a new Typesense client wrapper.
func NewTypesenseClient(url, apiKey, collection string) (*TypesenseClient, error) {
	if url == "" {
		return nil, errors.New("Typesense URL is required")
	}
	if apiKey == "" {
		return nil, errors.New("Typesense API key is required")
	}
	if collection == "" {
		return nil, errors.New("Typesense collection name is required")
	}

	return &TypesenseClient{
		url:        url,
		apiKey:     apiKey,
		collection: collection,
		batchSize:  defaultBatchSize,
		httpClient: &http.Client{},
	}, nil
}

// EnsureCollection creates the collection schema if it doesn't exist.
func (c *TypesenseClient) EnsureCollection(ctx context.Context) error {
	// Check if collection exists
	req, err := http.NewRequestWithContext(ctx, "GET", c.url+"/collections/"+c.collection, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-TYPESENSE-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("checking collection: %w", err)
	}
	defer resp.Body.Close()

	// Collection exists
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// Collection doesn't exist, create it
	if resp.StatusCode == http.StatusNotFound {
		return c.createCollection(ctx)
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}

func (c *TypesenseClient) createCollection(ctx context.Context) error {
	schema := map[string]interface{}{
		"name": c.collection,
		"fields": []map[string]interface{}{
			{"name": "id", "type": "string"},
			{"name": "file_path", "type": "string", "facet": true},
			{"name": "project_path", "type": "string", "facet": true},
			{"name": "project_type", "type": "string", "facet": true},
			{"name": "language", "type": "string", "facet": true},
			{"name": "chunk_type", "type": "string", "facet": true},
			{"name": "content", "type": "string"},
			{"name": "embedding", "type": "float[]", "num_dim": 768},
			{"name": "start_line", "type": "int32"},
			{"name": "end_line", "type": "int32"},
			{"name": "last_indexed", "type": "int64"},
		},
	}

	body, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("marshaling schema: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url+"/collections", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-TYPESENSE-API-KEY", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("creating collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: %s", string(respBody))
	}

	return nil
}

// UpsertChunks inserts or updates chunks in batches.
func (c *TypesenseClient) UpsertChunks(ctx context.Context, chunks []IndexedChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	batchSize := c.batchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[i:end]

		if err := c.upsertBatch(ctx, batch); err != nil {
			return fmt.Errorf("upserting batch %d: %w", i/batchSize, err)
		}
	}

	return nil
}

func (c *TypesenseClient) upsertBatch(ctx context.Context, chunks []IndexedChunk) error {
	// Build JSONL body
	var buf bytes.Buffer
	for _, chunk := range chunks {
		data, err := json.Marshal(chunk)
		if err != nil {
			return fmt.Errorf("marshaling chunk: %w", err)
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}

	endpoint := fmt.Sprintf("%s/collections/%s/documents/import?action=upsert", c.url, c.collection)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, &buf)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-TYPESENSE-API-KEY", c.apiKey)
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("importing documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("import failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Search performs hybrid search with both text query and vector embedding.
func (c *TypesenseClient) Search(ctx context.Context, query string, embedding []float32, limit int) ([]IndexedChunk, error) {
	searchRequest := map[string]interface{}{
		"searches": []map[string]interface{}{
			{
				"collection": c.collection,
				"q":          query,
				"query_by":   "content",
				"per_page":   limit,
			},
		},
	}

	// Add vector search if embedding provided
	if len(embedding) > 0 {
		searchRequest["searches"].([]map[string]interface{})[0]["vector_query"] = fmt.Sprintf("embedding:(%v)", formatEmbedding(embedding))
	}

	body, err := json.Marshal(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("marshaling search request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url+"/multi_search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-TYPESENSE-API-KEY", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var searchResp struct {
		Results []struct {
			Hits []struct {
				Document IndexedChunk `json:"document"`
			} `json:"hits"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var results []IndexedChunk
	if len(searchResp.Results) > 0 {
		for _, hit := range searchResp.Results[0].Hits {
			results = append(results, hit.Document)
		}
	}

	return results, nil
}

func formatEmbedding(embedding []float32) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, v := range embedding {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(&buf, "%f", v)
	}
	buf.WriteByte(']')
	return buf.String()
}

// DeleteByPath removes all documents for a given file path.
func (c *TypesenseClient) DeleteByPath(ctx context.Context, filePath string) error {
	if filePath == "" {
		return errors.New("file path is required")
	}

	filterBy := fmt.Sprintf("file_path:=%s", filePath)
	endpoint := fmt.Sprintf("%s/collections/%s/documents?filter_by=%s", c.url, c.collection, url.QueryEscape(filterBy))

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-TYPESENSE-API-KEY", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("deleting documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
