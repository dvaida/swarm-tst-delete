# swarm-indexer

A CLI tool that indexes text files into Typesense for AI context retrieval (RAG), with semantic chunking and Gemini embeddings.

## Features

- **Recursive directory indexing** with .gitignore support
- **Software project detection** (Go, Node, Python, Rust, Java, Ruby)
- **Semantic chunking** - code split by functions, docs by sections
- **Secrets protection** - skip secret files, redact inline secrets
- **Hybrid search** - Typesense text search + Gemini vector embeddings
- **Incremental updates** - hash-based change detection
- **High scale** - worker pool for 100k+ files

## Installation

```bash
go install github.com/dvaida/swarm-indexer/cmd/swarm-indexer@latest
```

Or build from source:
```bash
git clone https://github.com/dvaida/swarm-indexer.git
cd swarm-indexer
make build
```

## Usage

```bash
# Index one or more paths
swarm-indexer index /path/to/projects /path/to/docs

# Search indexed content
swarm-indexer search "authentication middleware"

# Check indexing status
swarm-indexer status
```

## Configuration

All configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `TYPESENSE_URL` | `http://localhost:8108` | Typesense server URL |
| `TYPESENSE_API_KEY` | (required) | Typesense API key |
| `TYPESENSE_COLLECTION` | `swarm-index` | Collection name |
| `GEMINI_API_KEY` | (required) | Google Gemini API key |
| `GEMINI_MODEL` | `gemini-embedding-001` | Embedding model |
| `GEMINI_RATE_LIMIT` | `60` | Requests per minute |
| `SWARM_INDEXER_WORKERS` | `8` | Parallel workers |
| `SWARM_INDEXER_BATCH_SIZE` | `100` | Typesense batch size |
| `SWARM_INDEXER_SKIP_FILES` | `.env,.setenv,*.pem,*.key,credentials.*` | Files to skip entirely |

## Requirements

- Go 1.23+
- Typesense server (local or cloud)
- Google Gemini API access

## How It Works

1. **Walk** directories recursively, respecting .gitignore
2. **Detect** if directory is a software project
3. **Filter** binary files and secret files
4. **Chunk** text files semantically (functions for code, sections for docs)
5. **Redact** any inline secrets found via Gitleaks
6. **Embed** chunks using Gemini API
7. **Index** into Typesense with hybrid search schema
8. **Store** metadata for incremental updates

## License

MIT
