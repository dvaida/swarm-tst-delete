# Issue #18: Bootstrap - Initialize Project Structure and Build Infrastructure

## Overview
Set up the foundational Go project structure for swarm-indexer CLI tool that all other issues depend on.

## Integration Tests to Write

### 1. CLI Integration Tests (`cmd/swarm-indexer/main_test.go`)
- Test that `--help` flag shows help text with three subcommands (index, search, status)
- Test that `index --help` shows index subcommand help
- Test that `search --help` shows search subcommand help
- Test that `status --help` shows status subcommand help
- Test running `index` command (stub should work, returns 0)
- Test running `search` command (stub should work, returns 0)
- Test running `status` command (stub should work, returns 0)

### 2. Config Integration Tests (`internal/config/config_test.go`)
- Test loading config with all defaults (when required vars set)
- Test TYPESENSE_URL default value
- Test TYPESENSE_COLLECTION default value
- Test GEMINI_MODEL default value
- Test GEMINI_RATE_LIMIT default value
- Test SWARM_INDEXER_WORKERS default value
- Test SWARM_INDEXER_BATCH_SIZE default value
- Test SWARM_INDEXER_SKIP_FILES default value
- Test loading custom values from ENV vars
- Test validation fails when TYPESENSE_API_KEY is missing
- Test validation fails when GEMINI_API_KEY is missing

## Implementation Approach

### Files to Create
1. `go.mod` - Go module definition
2. `cmd/swarm-indexer/main.go` - CLI entry point with cobra
3. `internal/config/config.go` - Configuration loading
4. `Makefile` - Build, test, lint targets
5. `.gitignore` - Go project gitignore
6. `.github/workflows/ci.yml` - CI workflow

### Directory Structure
```
swarm-indexer/
├── cmd/
│   └── swarm-indexer/
│       ├── main.go
│       └── main_test.go
├── internal/
│   └── config/
│       ├── config.go
│       └── config_test.go
├── plans/
│   └── issue-18.md
├── .github/
│   └── workflows/
│       └── ci.yml
├── go.mod
├── go.sum
├── Makefile
└── .gitignore
```

### Implementation Steps

1. Initialize go.mod
2. Write integration tests (CLI and config)
3. Create stub implementations (panic/error)
4. Run tests - verify all fail
5. Implement config.go with ENV loading
6. Implement main.go with cobra CLI
7. Create Makefile, .gitignore, CI workflow
8. Run tests - verify all pass
9. Verify `make build`, `make test`, `make lint` work
10. Commit changes

## Acceptance Criteria Verification
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` runs and passes
- [ ] `./bin/swarm-indexer --help` shows help with three subcommands
- [ ] Config loads from ENV vars with defaults
- [ ] `make build`, `make test`, `make lint` work
