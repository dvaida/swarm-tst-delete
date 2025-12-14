# Plan: Issue #10 - Incremental Sync

## Overview
Implement an incremental sync system that only reindexes changed files and removes deleted files from a Typesense index. This avoids redundant indexing by comparing file modification times against indexed timestamps.

## Types Required

### FileInfo
Represents a file to be potentially indexed:
- `Path` - file path
- `LastModified` - file mtime as Unix timestamp

### Project
Represents a registered project/root:
- `ID` - project identifier
- `RootPath` - base path for the project

### Document
Represents a document in Typesense:
- `ID` - computed from file path
- `Path` - file path
- `IndexedAt` - Unix timestamp when indexed

### SyncResult
Statistics about the sync operation:
- `Added` - count of newly indexed files
- `Updated` - count of reindexed files
- `Unchanged` - count of skipped files
- `Deleted` - count of removed documents
- `Failed` - count of failed operations

### Client (Interface)
Typesense client interface for testability:
- `GetDocument(id string) (*Document, error)` - fetch existing doc
- `UpsertDocument(doc Document) error` - add/update document
- `DeleteDocument(id string) error` - remove document
- `SearchByPathPrefix(prefix string) ([]Document, error)` - find docs by path prefix

## Integration Tests to Write

1. **Test_Sync_NewFiles** - Files not in index should be added
2. **Test_Sync_ModifiedFiles** - Files with mtime > indexed_at should be updated
3. **Test_Sync_UnchangedFiles** - Files with mtime <= indexed_at should be skipped
4. **Test_Sync_DeletedFiles** - Files no longer existing should be removed from index
5. **Test_Sync_MixedScenario** - Combination of new, modified, unchanged, and deleted
6. **Test_Sync_EmptyFileList** - Handle empty input gracefully
7. **Test_Sync_ClientErrors** - Count failures when client operations fail

## Implementation Approach

### Step 1: Setup Go Module
Initialize Go module and create package structure.

### Step 2: Define Types
Create all type definitions in `internal/sync/types.go`.

### Step 3: Implement ID Generation
Create helper to compute document ID from file path (hash-based).

### Step 4: Implement Sync Function
Main `Sync` function logic:
```
For each file:
  1. Compute document ID from path
  2. Fetch existing document from client
  3. If not exists: upsert and count as Added
  4. If exists and lastModified > indexedAt: upsert and count as Updated
  5. If exists and lastModified <= indexedAt: count as Unchanged
  6. On error: count as Failed

For cleanup:
  1. For each project, search docs by root path prefix
  2. Build set of current file paths
  3. Delete docs whose paths are not in current file set
  4. Count as Deleted or Failed
```

### Step 5: Return SyncResult
Aggregate all counts into result struct.

## File Structure
```
go.mod
internal/
  sync/
    types.go       # Type definitions
    sync.go        # Main Sync function
    sync_test.go   # Integration tests
    id.go          # Document ID generation
```

## Technical Notes
- Use crypto/sha256 for consistent document ID generation from paths
- Client is an interface for testability (mock implementation in tests)
- Process files sequentially for simplicity (parallel optimization can come later)
- Batch deletes would improve performance but add complexity - start simple
