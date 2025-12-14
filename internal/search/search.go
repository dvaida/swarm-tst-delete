package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// SearchResult represents a single search result
type SearchResult struct {
	FilePath    string  `json:"file_path"`
	ProjectPath string  `json:"project_path"`
	Language    string  `json:"language"`
	ChunkType   string  `json:"chunk_type"`
	Content     string  `json:"content"`
	StartLine   int     `json:"start_line"`
	EndLine     int     `json:"end_line"`
	Score       float64 `json:"score"`
}

// Searcher interface for performing searches
type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
	IsEmpty(ctx context.Context) (bool, error)
}

// MockSearcher is a mock implementation for testing
type MockSearcher struct {
	Results    []SearchResult
	EmptyIndex bool
	Err        error
}

func (m *MockSearcher) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if limit > 0 && limit < len(m.Results) {
		return m.Results[:limit], nil
	}
	return m.Results, nil
}

func (m *MockSearcher) IsEmpty(ctx context.Context) (bool, error) {
	return m.EmptyIndex, nil
}

// Search performs a hybrid search using the provided searcher
func Search(ctx context.Context, searcher Searcher, query string, limit int) ([]SearchResult, error) {
	return searcher.Search(ctx, query, limit)
}

// FormatResults formats search results as text or JSON
func FormatResults(results []SearchResult, asJSON bool) string {
	if asJSON {
		data, _ := json.MarshalIndent(results, "", "  ")
		return string(data)
	}

	if len(results) == 0 {
		return "No results found."
	}

	var sb strings.Builder
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("[%d] %s:%d-%d (%s) score: %.2f\n",
			i+1, r.FilePath, r.StartLine, r.EndLine, r.ChunkType, r.Score))

		content := r.Content
		const maxLen = 200
		if len(content) > maxLen {
			content = content[:maxLen] + "..."
		}
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			sb.WriteString("    " + line + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
