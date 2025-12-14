# swarm-indexer - Development Guide

## Project Overview

A Go CLI tool that indexes text files from registered paths into Typesense for AI context retrieval (RAG), using semantic chunking and Gemini embeddings for hybrid search.

## Architecture

```
swarm-indexer/
├── cmd/swarm-indexer/main.go        # CLI entry point (cobra)
├── internal/
│   ├── config/config.go             # ENV var loading + validation
│   ├── walker/
│   │   ├── walker.go                # Directory traversal with .gitignore
│   │   └── binary.go                # Binary file detection
│   ├── detector/
│   │   ├── project.go               # Software project detection
│   │   └── language.go              # Language detection per file
│   ├── metadata/metadata.go         # .swarm-indexer-metadata.json R/W
│   ├── secrets/
│   │   ├── scanner.go               # Gitleaks integration
│   │   └── redactor.go              # Inline secret redaction
│   ├── chunker/
│   │   ├── chunker.go               # Chunking orchestration
│   │   ├── code.go                  # Code-aware chunking
│   │   └── text.go                  # Text/docs chunking
│   ├── embeddings/gemini.go         # Gemini API client + rate limiting
│   ├── indexer/
│   │   ├── indexer.go               # Main orchestration + worker pool
│   │   └── typesense.go             # Typesense client wrapper
│   └── search/search.go             # Search + result formatting
├── go.mod
└── go.sum
```

## Development Workflow

### Prerequisites
- Go 1.23+ (latest stable)
- Typesense running locally (Docker)
- Gemini API key

### Build & Run
```bash
# Build
make build

# Run
./bin/swarm-indexer index /path/to/code
./bin/swarm-indexer search "query"
./bin/swarm-indexer status
```

### Testing
```bash
make test        # Run all tests
make lint        # Run linter
make build       # Build binary
```

### Environment Variables
All configuration via ENV vars with sensible defaults:

```bash
# Typesense (required: API key)
TYPESENSE_URL=http://localhost:8108      # default
TYPESENSE_API_KEY=                        # required
TYPESENSE_COLLECTION=swarm-index         # default

# Gemini (required: API key)
GEMINI_API_KEY=                           # required
GEMINI_MODEL=gemini-embedding-001        # default
GEMINI_RATE_LIMIT=60                     # default, requests/min

# Worker settings
SWARM_INDEXER_WORKERS=8                  # default
SWARM_INDEXER_BATCH_SIZE=100             # default

# Secrets (comma-separated patterns to skip entirely)
SWARM_INDEXER_SKIP_FILES=.env,.setenv,*.pem,*.key,credentials.*
```

## Code Style

- **Simplicity first**: Write minimal code that solves the problem
- **No over-engineering**: Avoid premature abstractions
- **Clear naming**: Functions and variables should be self-documenting
- **Error handling**: Return errors, don't panic. Log at appropriate levels.
- **Testing**: Write tests for public functions. Integration tests preferred.

## Tech Stack Decisions

| Choice | Rationale |
|--------|-----------|
| Go | Performance for 100k+ files, good concurrency primitives |
| Cobra | Standard CLI library, well-documented |
| Typesense Go v3 | Official client with circuit breaker support |
| Gitleaks | Mature secrets detection, can use as library |
| Gemini embeddings | Good quality, configurable model |

## Key Data Structures

### IndexedChunk (Typesense document)
```go
type IndexedChunk struct {
    ID            string    `json:"id"`
    FilePath      string    `json:"file_path"`
    ProjectPath   string    `json:"project_path"`
    ProjectType   string    `json:"project_type"`
    Language      string    `json:"language"`
    ChunkType     string    `json:"chunk_type"`
    Content       string    `json:"content"`
    Embedding     []float32 `json:"embedding"`
    StartLine     int       `json:"start_line"`
    EndLine       int       `json:"end_line"`
    LastIndexed   int64     `json:"last_indexed"`
}
```

### Metadata file (.swarm-indexer-metadata.json)
```go
type Metadata struct {
    LastIndexed  int64             `json:"last_indexed"`
    FileCount    int               `json:"file_count"`
    ContentHash  string            `json:"content_hash"`
    ProjectType  string            `json:"project_type"`
    Languages    []string          `json:"languages"`
    Dependencies map[string]string `json:"dependencies"`
}
```
