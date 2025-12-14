# Initial Requirements

## Original Request
Build a Go project that:
- Gets a list of registered paths
- Recursively lists folders/files
- Detects if a folder is a software project
- Creates metadata stored in `.swarm-indexer-metadata.json` per directory
- Indexes text documents into Typesense, avoiding private data (passwords, secrets)

## Clarified Requirements

### Configuration
- **All settings via ENV vars** with sensible defaults (no config files)
- Paths to index provided as **CLI arguments**, not config

### Path Registration & Metadata
- One `.swarm-indexer-metadata.json` per registered parent path
- Metadata includes: last indexed timestamp, file count, content hash, project type, languages, dependencies

### Software Project Detection (best effort)
- Package managers: go.mod, package.json, requirements.txt, pyproject.toml, Cargo.toml, pom.xml, build.gradle, Gemfile
- VCS: .git, .svn, .hg
- IDE configs: .vscode, .idea, .editorconfig
- Custom markers: not required for MVP

### Text Indexing
- Index **all text files** (not just code)
- Skip binary files (detect via null bytes in first 8KB)
- Respect .gitignore at all directory levels

### Secrets Handling (two types)
- **Type A - Skip entirely**: .env, .setenv, *.pem, *.key, credentials.* (configurable via ENV)
- **Type B - Redact inline**: Files with detected secrets get indexed but secrets replaced with `[REDACTED]`
- Use **Gitleaks** for secrets detection

### Chunking Strategy
- **Semantic chunking** for RAG retrieval
- Code: Split by functions/methods/classes
- Docs: Split by paragraphs/headers/sections
- Config: Split by top-level keys

### Search & Embeddings
- **Typesense hybrid search** (text + vector)
- **Gemini embeddings** (gemini-embedding-001, configurable)
- Rate limiting for Gemini API

### Incremental Updates
- **Hash-based** change detection
- Only re-index modified files

### Scale
- Target: **100k+ files**
- Worker pool parallelization
- Batch Typesense uploads
- Gemini rate limiting

### CLI Commands (minimal)
```
swarm-indexer index <paths...>    # Index specified paths
swarm-indexer search <query>      # Search indexed content
swarm-indexer status              # Show indexing status
```

## Non-Functional Requirements

- **Language**: Go (latest stable, 1.23+)
- **Performance**: Handle large codebases efficiently
- **Simplicity**: Minimal dependencies, clear code structure
- **Error handling**: Graceful degradation, clear error messages

## Constraints & Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| CLI framework | Cobra | Standard, well-documented |
| Typesense client | Official Go v3 | Circuit breaker support |
| Secrets detection | Gitleaks | Mature, can use as library |
| Embedding model | Gemini | Good quality, configurable |
| Project structure | cmd/ + internal/ | Go standard layout |
