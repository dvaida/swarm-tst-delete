# Plan for Issue #27: Implement search command with formatted output

## Overview
Implement the `search` CLI command that performs hybrid search (text + vector) against Typesense and displays formatted results. This includes:
- CLI wiring with cobra
- Calling Typesense hybrid search
- Generating query embeddings via Gemini
- Formatting results in text or JSON
- Pagination support via --limit flag

## Files to Create
1. `go.mod` - Go module definition
2. `internal/config/config.go` - Environment configuration
3. `internal/embeddings/gemini.go` - Gemini API client for query embedding
4. `internal/indexer/typesense.go` - Typesense client wrapper
5. `internal/search/search.go` - Search logic and result formatting
6. `cmd/swarm-indexer/main.go` - CLI entry point with search command

## Integration Tests to Write

### 1. TestSearch_ReturnsResults
- Setup: Mock Typesense with test data
- Execute search with query
- Verify results returned with correct fields

### 2. TestSearch_WithLimit
- Execute search with --limit flag
- Verify correct number of results returned

### 3. TestSearch_NoResults
- Search for query with no matches
- Verify empty results handled gracefully

### 4. TestSearch_EmptyIndex
- Search when index is empty
- Verify helpful message displayed

### 5. TestFormatResults_TextFormat
- Format results as text (default)
- Verify output matches expected format with file path, lines, snippet, score

### 6. TestFormatResults_JSONFormat
- Format results as JSON
- Verify valid JSON with correct structure

### 7. TestSearchCommand_Integration
- Test full CLI command execution
- Verify search runs and outputs correctly

## Implementation Approach

1. **Config** - Load Typesense and Gemini settings from env vars
2. **Gemini Client** - Generate embeddings for search query
3. **Typesense Client** - Execute hybrid search (text + vector)
4. **Search** - Orchestrate embedding + search + format
5. **CLI** - Wire cobra command with flags

## Dependencies
- github.com/spf13/cobra - CLI framework
- github.com/typesense/typesense-go/v3 - Typesense client
- google.golang.org/genai - Gemini SDK (or HTTP client)

## Acceptance Criteria
- `swarm-indexer search "query"` returns relevant results
- Results show file path, lines, snippet preview
- JSON output available with --json flag
- Respects --limit flag (default 10)
- Handles no results gracefully
- Shows helpful message if index is empty
