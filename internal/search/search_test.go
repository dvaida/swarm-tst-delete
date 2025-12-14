package search_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/dvaida/swarm-indexer/internal/search"
)

// TestSearch_ReturnsResults tests that search returns results for a matching query
func TestSearch_ReturnsResults(t *testing.T) {
	ctx := context.Background()

	// Create a mock searcher with test data
	mockSearcher := &search.MockSearcher{
		Results: []search.SearchResult{
			{
				FilePath:    "src/auth/middleware.go",
				ProjectPath: "/path/to/project",
				Language:    "go",
				ChunkType:   "function",
				Content:     "func AuthMiddleware(next http.Handler) http.Handler {\n    // Validates JWT token\n}",
				StartLine:   45,
				EndLine:     67,
				Score:       0.92,
			},
		},
	}

	results, err := search.Search(ctx, mockSearcher, "authentication middleware", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.FilePath != "src/auth/middleware.go" {
		t.Errorf("expected file path 'src/auth/middleware.go', got %q", result.FilePath)
	}
	if result.Score != 0.92 {
		t.Errorf("expected score 0.92, got %f", result.Score)
	}
	if result.StartLine != 45 {
		t.Errorf("expected start line 45, got %d", result.StartLine)
	}
	if result.EndLine != 67 {
		t.Errorf("expected end line 67, got %d", result.EndLine)
	}
}

// TestSearch_WithLimit tests that search respects the limit parameter
func TestSearch_WithLimit(t *testing.T) {
	ctx := context.Background()

	mockSearcher := &search.MockSearcher{
		Results: []search.SearchResult{
			{FilePath: "file1.go", Score: 0.9},
			{FilePath: "file2.go", Score: 0.8},
			{FilePath: "file3.go", Score: 0.7},
			{FilePath: "file4.go", Score: 0.6},
			{FilePath: "file5.go", Score: 0.5},
		},
	}

	results, err := search.Search(ctx, mockSearcher, "test query", 3)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results (limit=3), got %d", len(results))
	}
}

// TestSearch_NoResults tests graceful handling when no results match
func TestSearch_NoResults(t *testing.T) {
	ctx := context.Background()

	mockSearcher := &search.MockSearcher{
		Results: []search.SearchResult{},
	}

	results, err := search.Search(ctx, mockSearcher, "nonexistent query xyz", 10)
	if err != nil {
		t.Fatalf("Search should not error on no results: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

// TestSearch_EmptyIndex tests handling when the index is empty
func TestSearch_EmptyIndex(t *testing.T) {
	ctx := context.Background()

	mockSearcher := &search.MockSearcher{
		EmptyIndex: true,
		Results:    []search.SearchResult{},
	}

	results, err := search.Search(ctx, mockSearcher, "any query", 10)
	if err != nil {
		t.Fatalf("Search should not error on empty index: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("expected 0 results for empty index, got %d", len(results))
	}
}

// TestFormatResults_TextFormat tests text output formatting
func TestFormatResults_TextFormat(t *testing.T) {
	results := []search.SearchResult{
		{
			FilePath:    "src/auth/middleware.go",
			ProjectPath: "/path/to/project",
			Language:    "go",
			ChunkType:   "function",
			Content:     "func AuthMiddleware(next http.Handler) http.Handler {\n    // Validates JWT token and sets user context\n}",
			StartLine:   45,
			EndLine:     67,
			Score:       0.92,
		},
		{
			FilePath:    "src/handlers/login.go",
			ProjectPath: "/path/to/project",
			Language:    "go",
			ChunkType:   "function",
			Content:     "func HandleLogin(w http.ResponseWriter, r *http.Request) {\n    // Handle login\n}",
			StartLine:   23,
			EndLine:     41,
			Score:       0.87,
		},
	}

	output := search.FormatResults(results, false)

	// Check that output contains expected elements
	if !strings.Contains(output, "[1]") {
		t.Error("expected output to contain '[1]' result number")
	}
	if !strings.Contains(output, "[2]") {
		t.Error("expected output to contain '[2]' result number")
	}
	if !strings.Contains(output, "src/auth/middleware.go:45-67") {
		t.Error("expected output to contain file path with line numbers")
	}
	if !strings.Contains(output, "(function)") {
		t.Error("expected output to contain chunk type")
	}
	if !strings.Contains(output, "score: 0.92") {
		t.Error("expected output to contain score")
	}
	if !strings.Contains(output, "func AuthMiddleware") {
		t.Error("expected output to contain code snippet")
	}
}

// TestFormatResults_JSONFormat tests JSON output formatting
func TestFormatResults_JSONFormat(t *testing.T) {
	results := []search.SearchResult{
		{
			FilePath:    "src/auth/middleware.go",
			ProjectPath: "/path/to/project",
			Language:    "go",
			ChunkType:   "function",
			Content:     "func AuthMiddleware(next http.Handler) http.Handler {}",
			StartLine:   45,
			EndLine:     67,
			Score:       0.92,
		},
	}

	output := search.FormatResults(results, true)

	// Verify it's valid JSON
	var parsed []search.SearchResult
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 result in JSON, got %d", len(parsed))
	}

	if parsed[0].FilePath != "src/auth/middleware.go" {
		t.Errorf("expected file_path 'src/auth/middleware.go', got %q", parsed[0].FilePath)
	}
	if parsed[0].Score != 0.92 {
		t.Errorf("expected score 0.92, got %f", parsed[0].Score)
	}
}

// TestFormatResults_EmptyResults tests formatting of empty results
func TestFormatResults_EmptyResults(t *testing.T) {
	results := []search.SearchResult{}

	// Text format
	textOutput := search.FormatResults(results, false)
	if !strings.Contains(textOutput, "No results found") {
		t.Error("expected 'No results found' message in text output")
	}

	// JSON format
	jsonOutput := search.FormatResults(results, true)
	var parsed []search.SearchResult
	if err := json.Unmarshal([]byte(jsonOutput), &parsed); err != nil {
		t.Fatalf("empty JSON output is not valid: %v", err)
	}
	if len(parsed) != 0 {
		t.Errorf("expected empty JSON array, got %d elements", len(parsed))
	}
}

// TestFormatResults_TruncatesLongContent tests that long content is truncated
func TestFormatResults_TruncatesLongContent(t *testing.T) {
	longContent := strings.Repeat("x", 1000)
	results := []search.SearchResult{
		{
			FilePath:  "test.go",
			Content:   longContent,
			StartLine: 1,
			EndLine:   100,
			Score:     0.5,
		},
	}

	output := search.FormatResults(results, false)

	// Text output should truncate content and add ellipsis
	if strings.Contains(output, longContent) {
		t.Error("expected long content to be truncated")
	}
	if !strings.Contains(output, "...") {
		t.Error("expected truncated content to have '...'")
	}
}
