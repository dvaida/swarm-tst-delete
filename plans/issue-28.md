# Issue 28: Implement Status Command

## Overview

Implement the `status` CLI command that shows indexing state for specified paths. This command reads `.swarm-indexer-metadata.json` files and queries Typesense for collection stats.

## Design Decision

Per issue notes, using option 2: paths as arguments (`swarm-indexer status /path/one /path/two`). This is consistent with the `index` command.

## Integration Tests to Write

1. **TestStatusCommand_ShowsPathMetadata** - When given a path with valid metadata, displays path info (type, files, languages, last indexed)
2. **TestStatusCommand_DetectsChanges** - When content hash changed, shows "Changes detected" status
3. **TestStatusCommand_ShowsUpToDate** - When content hash matches, shows "Up to date" status
4. **TestStatusCommand_MissingMetadata** - When metadata file doesn't exist, handles gracefully with message
5. **TestStatusCommand_TypesenseStats** - Displays Typesense collection document count
6. **TestStatusCommand_TypesenseConnectionError** - Handles Typesense connection errors gracefully
7. **TestStatusCommand_MultiplePaths** - Shows status for multiple paths

## Implementation Approach

### Files to Create/Modify

1. `go.mod` - Initialize Go module (project is empty)
2. `cmd/swarm-indexer/main.go` - CLI entry point with status subcommand
3. `internal/config/config.go` - ENV var loading for Typesense settings
4. `internal/metadata/metadata.go` - Metadata file reading
5. `internal/indexer/typesense.go` - Typesense client for collection stats
6. `internal/status/status.go` - Status command logic

### Key Functions

```go
// internal/metadata/metadata.go
type Metadata struct {
    LastIndexed  int64             `json:"last_indexed"`
    FileCount    int               `json:"file_count"`
    ContentHash  string            `json:"content_hash"`
    ProjectType  string            `json:"project_type"`
    Languages    []string          `json:"languages"`
    Dependencies map[string]string `json:"dependencies"`
}

func Load(path string) (*Metadata, error)
func ComputeContentHash(path string) (string, error)

// internal/indexer/typesense.go
type Client struct { ... }
func NewClient(url, apiKey, collection string) (*Client, error)
func (c *Client) GetCollectionStats() (*CollectionStats, error)

// internal/status/status.go
func Run(paths []string, tsClient *typesense.Client) error
```

### Output Format (from issue)

```
Indexed Paths:
─────────────────────────────────────────────────────────────

📁 /Users/me/projects/webapp
   Type: node | Files: 1,234 | Languages: ts, js, json
   Last indexed: 2024-01-15 14:30:22
   Status: ✓ Up to date

📁 /Users/me/projects/backend
   Type: go | Files: 567 | Languages: go, yaml
   Last indexed: 2024-01-14 09:15:00
   Status: ⚠ Changes detected (re-index needed)

─────────────────────────────────────────────────────────────
Typesense Collection: swarm-index
   Documents: 45,678
   URL: http://localhost:8108
```

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/typesense/typesense-go/v3` - Typesense client

## Steps

1. Initialize Go module
2. Create test files with integration tests
3. Create stub implementations that panic
4. Verify all tests fail
5. Implement minimal code to pass tests
6. Verify all tests pass
