# Issue #10: Incremental Sync

## Overview
Implement an incremental sync package that only reindexes changed files and removes deleted files from Typesense.

## Core Types

### Input Types
- `FileInfo` - Represents a file to sync with path and last modified time
- `Project` - Represents a project with root path for filtering
- `Client` - Interface for Typesense operations (get, upsert, delete, search)

### Output Type
- `SyncResult` - Contains counts: Added, Updated, Unchanged, Deleted, Failed

## Integration Tests to Write

1. **TestSync_NewFiles** - Files not in Typesense get added (Added count)
2. **TestSync_UpdatedFiles** - Files with mtime > indexed_at get updated (Updated count)
3. **TestSync_UnchangedFiles** - Files with mtime <= indexed_at are skipped (Unchanged count)
4. **TestSync_DeletedFiles** - Files in Typesense but not in file list get deleted (Deleted count)
5. **TestSync_FailedOperations** - Track failures in Failed count
6. **TestSync_MixedOperations** - Combination of add, update, skip, delete in one sync
7. **TestSync_EmptyFileList** - Syncing empty list deletes all existing docs
8. **TestSync_NoExistingDocs** - All files are new when Typesense is empty

## Implementation Approach

### Document ID
- Compute from file path (hash or base64 encoding for uniqueness)

### Sync Logic
For each file:
1. Compute document ID from path
2. Fetch existing document from Typesense
3. If not exists: index (Add)
4. If exists and `last_modified > indexed_at`: reindex (Update)
5. If exists and `last_modified <= indexed_at`: skip (Unchanged)

### Cleanup Logic
1. Search Typesense for all documents under registered project roots
2. Compare document paths against current file list
3. Delete documents whose files no longer exist

### Client Interface
```go
type Client interface {
    GetDocument(collection string, id string) (*Document, error)
    UpsertDocument(collection string, doc Document) error
    DeleteDocument(collection string, id string) error
    SearchDocuments(collection string, filter string) ([]Document, error)
}
```

### Document Structure
```go
type Document struct {
    ID         string `json:"id"`
    Path       string `json:"path"`
    IndexedAt  int64  `json:"indexed_at"`  // Unix timestamp
    Content    string `json:"content"`
    // ... other fields as needed
}
```

## File Structure
```
internal/
  sync/
    sync.go        # Main implementation
    sync_test.go   # Integration tests
    types.go       # Type definitions
```
