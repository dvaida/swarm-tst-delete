package indexer

import "errors"

// ErrConnectionFailed indicates a connection error to Typesense
var ErrConnectionFailed = errors.New("typesense connection failed")

// CollectionStats contains statistics about a Typesense collection
type CollectionStats struct {
	DocumentCount  int64
	CollectionName string
}

// TypesenseClient defines the interface for Typesense operations
type TypesenseClient interface {
	GetCollectionStats() (*CollectionStats, error)
	GetURL() string
}

// Client is the real Typesense client implementation
type Client struct {
	url        string
	apiKey     string
	collection string
}

// NewClient creates a new Typesense client
func NewClient(url, apiKey, collection string) (*Client, error) {
	return &Client{
		url:        url,
		apiKey:     apiKey,
		collection: collection,
	}, nil
}

// GetCollectionStats returns statistics about the collection
func (c *Client) GetCollectionStats() (*CollectionStats, error) {
	// Real implementation would query Typesense
	// For now, return placeholder (will be implemented in issue 7)
	return nil, ErrConnectionFailed
}

// GetURL returns the Typesense server URL
func (c *Client) GetURL() string {
	return c.url
}

// MockClient is a mock implementation for testing
type MockClient struct {
	Stats      *CollectionStats
	StatsError error
	URL        string
}

// GetCollectionStats returns mock stats or error
func (m *MockClient) GetCollectionStats() (*CollectionStats, error) {
	if m.StatsError != nil {
		return nil, m.StatsError
	}
	return m.Stats, nil
}

// GetURL returns the mock URL
func (m *MockClient) GetURL() string {
	if m.URL == "" {
		return "http://localhost:8108"
	}
	return m.URL
}
