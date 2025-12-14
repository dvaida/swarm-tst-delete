# Plan: Issue #25 - Implement Typesense Client Wrapper

## Overview

Implement a Typesense client wrapper for the swarm-indexer project. This wrapper will handle:
- Client initialization from ENV vars
- Collection schema creation with embedding vector field
- Batch document upserts
- Hybrid search (text + vector)
- Document deletion by file path

## Integration Tests to Write

1. **TestNewTypesenseClient_Success** - Valid configuration creates client
2. **TestNewTypesenseClient_EmptyAPIKey** - Returns error when API key is empty
3. **TestEnsureCollection_CreatesIfNotExists** - Collection is created with correct schema
4. **TestEnsureCollection_AlreadyExists** - No error when collection exists
5. **TestUpsertChunks_SingleChunk** - Can upsert a single chunk
6. **TestUpsertChunks_MultipleBatches** - Large batch is split according to batch size
7. **TestUpsertChunks_EmptySlice** - Empty slice doesn't cause error
8. **TestSearch_ReturnsResults** - Search returns matching documents
9. **TestSearch_HybridSearch** - Hybrid search with both query and embedding
10. **TestSearch_EmptyResults** - Returns empty slice when no matches
11. **TestDeleteByPath_RemovesDocuments** - Documents for path are deleted
12. **TestDeleteByPath_NoMatchingDocs** - No error when no documents match

## Implementation Approach

1. Create `go.mod` and `go.sum` with required dependencies
2. Create `internal/indexer/typesense.go` with:
   - `IndexedChunk` struct with JSON tags
   - `TypesenseClient` struct
   - `NewTypesenseClient()` constructor
   - `EnsureCollection()` method
   - `UpsertChunks()` method with batching
   - `Search()` method with hybrid search
   - `DeleteByPath()` method
3. Create `internal/config/config.go` for ENV var loading (needed for batch size)

## Files to Create

- `go.mod` - Module definition
- `internal/indexer/typesense.go` - Main implementation
- `internal/indexer/typesense_test.go` - Integration tests
- `internal/config/config.go` - Config loading (for batch size)

## Dependencies

- `github.com/typesense/typesense-go/v3` - Typesense Go client
