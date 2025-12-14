# Issue #26: Implement main indexer with worker pool

## Overview
Implement the main indexer orchestrator that coordinates the full indexing pipeline with parallel processing using a worker pool. This is the integration point that brings together all the foundation modules (walker, detector, secrets, chunker, embeddings, typesense, metadata).

## Key Requirements
1. **Worker Pool**: Configurable parallel processing via `SWARM_INDEXER_WORKERS` (default: 8)
2. **Batching**: Batch Typesense uploads via `SWARM_INDEXER_BATCH_SIZE` (default: 100)
3. **Incremental Indexing**: Skip unchanged directories using metadata content hash
4. **Error Handling**: Continue on single file failures, log progress and errors
5. **Pipeline Flow**: walk → detect secrets → chunk → embed → upsert

## Dependencies
Since this is the first issue being implemented, I need to create stub/mock versions of the dependencies:
- `internal/config` - Configuration loading
- `internal/walker` - Directory traversal
- `internal/detector` - Project/language detection
- `internal/secrets` - Secret scanning and redaction
- `internal/chunker` - Semantic chunking
- `internal/embeddings` - Gemini embeddings
- `internal/metadata` - Metadata file R/W
- `internal/indexer/typesense.go` - Typesense client

## Integration Tests to Write

### 1. Test Worker Pool Respects Worker Count
- Create an indexer with specific worker count
- Verify concurrent processing happens with that count
- Track maximum concurrent workers

### 2. Test Full Pipeline Processing
- Set up a directory with test files
- Run indexer
- Verify files are processed through: walk → secrets → chunk → embed → upsert

### 3. Test Batch Processing
- Process files and verify embeddings/upserts are batched
- Verify batch size is respected

### 4. Test Incremental Indexing (Skip Unchanged)
- Index a directory
- Index again without changes
- Verify second run skips processing

### 5. Test Error Handling (Continue on Failures)
- Have some files that cause processing errors
- Verify indexer continues processing remaining files
- Verify errors are logged/returned appropriately

### 6. Test Progress Logging
- Index multiple files
- Verify progress is logged

### 7. Test Metadata Update
- Successfully index a directory
- Verify metadata file is updated with correct hash

## Implementation Approach

1. Create minimal stub interfaces for dependencies
2. Implement `Indexer` struct with configuration
3. Implement worker pool using Go channels and sync.WaitGroup
4. Implement the pipeline stages
5. Handle batching for embeddings and Typesense upserts
6. Implement incremental detection via content hash

## Files to Create
- `go.mod` - Module definition
- `internal/config/config.go` - Config struct (stub)
- `internal/walker/walker.go` - Walker interface (stub)
- `internal/detector/detector.go` - Detector interface (stub)
- `internal/secrets/scanner.go` - Scanner interface (stub)
- `internal/chunker/chunker.go` - Chunker interface (stub)
- `internal/embeddings/gemini.go` - Embeddings client interface (stub)
- `internal/metadata/metadata.go` - Metadata manager (stub)
- `internal/indexer/typesense.go` - Typesense client (stub)
- `internal/indexer/indexer.go` - Main indexer implementation
- `internal/indexer/indexer_test.go` - Integration tests
