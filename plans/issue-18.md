# Plan: Issue #18 - Bootstrap Project Structure

## Overview
Set up the foundational Go project structure for swarm-indexer. This includes:
- Go module initialization
- Directory structure
- Cobra CLI with subcommands
- Config package for ENV var loading
- Build infrastructure (Makefile, CI, gitignore)

## Integration Tests to Write

1. **Config Tests** (`internal/config/config_test.go`):
   - Test loading config with all ENV vars set
   - Test loading config with defaults (no ENV vars)
   - Test required field validation (TYPESENSE_API_KEY, GEMINI_API_KEY)
   - Test parsing of comma-separated SWARM_INDEXER_SKIP_FILES
   - Test integer parsing for rate limit, workers, batch size

2. **CLI Tests** (`cmd/swarm-indexer/main_test.go`):
   - Test that `--help` shows all three subcommands
   - Test that `index` subcommand exists
   - Test that `search` subcommand exists
   - Test that `status` subcommand exists

## Implementation Approach

### Step 1: Create directory structure
```
cmd/swarm-indexer/
internal/config/
internal/walker/
internal/detector/
internal/metadata/
internal/secrets/
internal/chunker/
internal/embeddings/
internal/indexer/
internal/search/
.github/workflows/
```

### Step 2: Initialize Go module
```bash
go mod init github.com/dvaida/swarm-indexer
```

### Step 3: Create files
- `internal/config/config.go` - Config struct and Load function
- `internal/config/config_test.go` - Tests for config
- `cmd/swarm-indexer/main.go` - Cobra CLI setup
- `cmd/swarm-indexer/main_test.go` - CLI tests
- `Makefile` - Build targets
- `.gitignore` - Go project ignores
- `.github/workflows/ci.yml` - CI workflow

## Files to Create

| File | Purpose |
|------|---------|
| `go.mod` | Go module definition |
| `cmd/swarm-indexer/main.go` | CLI entry point |
| `cmd/swarm-indexer/main_test.go` | CLI tests |
| `internal/config/config.go` | Config loading |
| `internal/config/config_test.go` | Config tests |
| `Makefile` | Build automation |
| `.gitignore` | Git ignores |
| `.github/workflows/ci.yml` | CI pipeline |

## Acceptance Criteria Verification
- `go build ./...` succeeds
- `go test ./...` runs and passes
- `./bin/swarm-indexer --help` shows help with three subcommands
- Config loads from ENV vars with defaults
- `make build`, `make test`, `make lint` work
