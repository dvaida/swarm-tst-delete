# Implementation Plan

## Phase Overview

The implementation is broken into 11 GitHub issues designed for parallel execution by AI agents.

## Issue Breakdown

### Issue 0: Bootstrap (runs first, alone)
**Initialize project structure and build infrastructure**

Sets up the foundational Go project that all other issues depend on:
- Go module initialization
- Directory structure
- Cobra CLI skeleton
- Config loading from ENV vars
- Makefile, .gitignore, CI workflow

### Issue 1: Directory Walker
**Implement directory walker with .gitignore support**

Files: `internal/walker/walker.go`, `internal/walker/binary.go`
- Recursive traversal
- .gitignore parsing
- Binary file detection
- Symlink safety

### Issue 2: Project Detector
**Implement software project detection**

Files: `internal/detector/project.go`, `internal/detector/language.go`
- Marker file detection (go.mod, package.json, etc.)
- VCS detection
- Language detection by extension
- Basic dependency extraction

### Issue 3: Metadata Manager
**Implement metadata storage**

Files: `internal/metadata/metadata.go`
- Load/Save `.swarm-indexer-metadata.json`
- Content hash computation
- Change detection

### Issue 4: Secrets Scanner
**Implement secrets detection with Gitleaks**

Files: `internal/secrets/scanner.go`, `internal/secrets/redactor.go`
- Gitleaks integration
- File skip patterns
- Inline secret redaction

### Issue 5: Semantic Chunker
**Implement semantic chunking**

Files: `internal/chunker/chunker.go`, `internal/chunker/code.go`, `internal/chunker/text.go`
- Code chunking by functions
- Text chunking by paragraphs/headers
- Line number tracking

### Issue 6: Gemini Client
**Implement Gemini embeddings with rate limiting**

Files: `internal/embeddings/gemini.go`
- Gemini API client
- Rate limiting
- Retry with backoff

### Issue 7: Typesense Client
**Implement Typesense client wrapper**

Files: `internal/indexer/typesense.go`
- Collection schema creation
- Batch upserts
- Hybrid search
- Delete by path

### Issue 8: Indexer Orchestration
**Implement main indexer with worker pool**

Files: `internal/indexer/indexer.go`
- Worker pool
- Full pipeline orchestration
- Progress tracking
- Incremental handling

### Issue 9: Search Command
**Implement search CLI command**

Files: `internal/search/search.go`, `cmd/swarm-indexer/main.go`
- Hybrid search execution
- Result formatting
- JSON output option

### Issue 10: Status Command
**Implement status CLI command**

Files: `cmd/swarm-indexer/main.go`
- Read metadata files
- Show collection stats
- Indicate stale indexes

## Parallelization Strategy

```
Bootstrap (Issue 0) - runs first, alone
         │
         ▼
    ┌────┴────┬─────────┬─────────┐
    │         │         │         │
Issue 1    Issue 2   Issue 3   Issue 4
(walker)   (detector) (metadata) (secrets)
    │         │         │         │
    └────┬────┴─────────┴─────────┘
         │
         ▼
    ┌────┴────┬─────────┐
    │         │         │
Issue 5    Issue 6   Issue 7
(chunker)  (gemini)  (typesense)
    │         │         │
    └────┬────┴─────────┘
         │
         ▼
      Issue 8
    (indexer orchestration)
         │
         ▼
    ┌────┴────┐
    │         │
Issue 9    Issue 10
(search)   (status)
```

**Wave 1** (after bootstrap): Issues 1, 2, 3, 4 - independent modules, different directories
**Wave 2**: Issues 5, 6, 7 - can run in parallel, depend on wave 1
**Wave 3**: Issue 8 - integration, depends on all above
**Wave 4**: Issues 9, 10 - final commands, can run in parallel

## File Ownership (Conflict Avoidance)

| Issue | Owns Files |
|-------|------------|
| 0 | cmd/*, go.mod, Makefile, .github/*, internal/config/* |
| 1 | internal/walker/* |
| 2 | internal/detector/* |
| 3 | internal/metadata/* |
| 4 | internal/secrets/* |
| 5 | internal/chunker/* |
| 6 | internal/embeddings/* |
| 7 | internal/indexer/typesense.go |
| 8 | internal/indexer/indexer.go |
| 9 | internal/search/*, updates cmd/ |
| 10 | updates cmd/ |

Issues 9 and 10 both update cmd/ but in different functions - can be merged sequentially.
