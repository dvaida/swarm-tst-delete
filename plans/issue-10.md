# Plan: Issue #10 - Incremental Sync

## Overview
Implement incremental sync functionality that:
1. Only reindexes files that have changed (based on mtime vs indexed_at)
2. Removes deleted files from the Typesense index
3. Tracks statistics about what was added, updated, unchanged, deleted, and failed

## Types Required

### Input Types
- `FileInfo`: Contains file path, last modified time, and content to index
- `Project`: Contains project path/root information
- `TypesenseClient`: Interface for Typesense operations (search, upsert, delete)

### Output Types
- `SyncResult`: Contains Added, Updated, Unchanged, Deleted, Failed counts

## Integration Tests to Write

1. **TestSyncNewFiles** - Files not in index should be added, result shows correct Added count
2. **TestSyncUpdatedFiles** - Files with mtime > indexed_at should be reindexed, result shows correct Updated count
3. **TestSyncUnchangedFiles** - Files with mtime <= indexed_at should be skipped, result shows correct Unchanged count
4. **TestSyncDeletedFiles** - Files in index but not in scan should be deleted, result shows correct Deleted count
5. **TestSyncMixedOperations** - Mix of new, updated, unchanged, and deleted files
6. **TestSyncEmpty** - Empty file list with empty index returns all zeros
7. **TestSyncFailedOperations** - When client operations fail, Failed count increments

## Implementation Approach

### Step 1: Define types in `internal/sync/sync.go`
- `FileInfo` struct with Path, LastModified, Content
- `Project` struct with Root path
- `TypesenseClient` interface with methods: GetDocument, UpsertDocument, DeleteDocument, SearchByPathPrefix
- `SyncResult` struct with Added, Updated, Unchanged, Deleted, Failed int fields
- `Document` struct representing indexed data with ID, Path, Content, IndexedAt

### Step 2: Implement `Sync` function
```
func Sync(files []FileInfo, projects []Project, client TypesenseClient) SyncResult
```

Logic:
1. For each file in files:
   - Compute document ID from path (hash or sanitized path)
   - Try to get existing document from client
   - If not exists: upsert new document, increment Added
   - If exists and file.LastModified > doc.IndexedAt: upsert, increment Updated
   - If exists and file.LastModified <= doc.IndexedAt: skip, increment Unchanged
   - On any error: increment Failed

2. For cleanup (deleted files):
   - For each project, get all document IDs with that path prefix
   - Build set of current file paths
   - Delete documents whose paths are not in current file set
   - Increment Deleted for each

### Step 3: Document ID Generation
- Use path-based ID: sanitize path to create valid Typesense document ID
- Replace `/` with `_`, remove special chars, etc.

## Files to Create
- `internal/sync/sync.go` - Types and Sync function
- `internal/sync/sync_test.go` - Integration tests

## Dependencies
- Standard library only (time, testing)
- No external Typesense library (use interface for abstraction)
