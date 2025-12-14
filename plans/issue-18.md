# Plan: Issue #18 - Bootstrap: Initialize project structure and build infrastructure

## Overview
Set up the foundational Go project structure that all other issues depend on. This includes:
- Go module initialization
- Directory structure creation
- Cobra CLI skeleton with stub subcommands
- Environment variable configuration loading
- Build infrastructure (Makefile, CI, .gitignore)

## Integration Tests to Write

### 1. Config Loading Tests (`internal/config/config_test.go`)
- Test that config loads default values when ENV vars not set
- Test that config loads from ENV vars when set
- Test that required vars (TYPESENSE_API_KEY, GEMINI_API_KEY) cause validation error when missing
- Test that SWARM_INDEXER_SKIP_FILES parses comma-separated patterns correctly

### 2. CLI Tests (`cmd/swarm-indexer/main_test.go`)
- Test that `--help` flag works and shows subcommands
- Test that `index` subcommand exists
- Test that `search` subcommand exists
- Test that `status` subcommand exists

## Implementation Approach

### Step 1: Initialize Go module
```bash
go mod init github.com/dvaida/swarm-indexer
```

### Step 2: Create directory structure
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

### Step 3: Write integration tests first
- Config tests in `internal/config/config_test.go`
- CLI tests in `cmd/swarm-indexer/main_test.go`

### Step 4: Create stub implementations
- `internal/config/config.go` - Config struct and Load() that panics
- `cmd/swarm-indexer/main.go` - Cobra root and subcommands that panic

### Step 5: Verify tests fail

### Step 6: Implement minimal code
- Config: Load ENV vars with defaults
- CLI: Working root command with index/search/status subcommands

### Step 7: Create build infrastructure
- Makefile with build, test, lint targets
- .gitignore for Go projects
- .github/workflows/ci.yml for CI

### Step 8: Verify all tests pass and acceptance criteria met

## Files to Create/Modify
1. `go.mod` - Module initialization
2. `internal/config/config.go` - Config loading
3. `internal/config/config_test.go` - Config tests
4. `cmd/swarm-indexer/main.go` - CLI entry point
5. `cmd/swarm-indexer/main_test.go` - CLI tests
6. `Makefile` - Build targets
7. `.gitignore` - Go project ignores
8. `.github/workflows/ci.yml` - CI workflow
