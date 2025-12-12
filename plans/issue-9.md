# Plan: Issue #9 - Document Indexing

## Overview
Implement the `IndexFiles` function that reads files and indexes them to Typesense. The function builds documents from file information, handles UTF-8 validation, and upserts to Typesense with graceful error handling.

## Integration Tests to Write

1. **TestIndexFilesSuccess** - Index multiple valid files successfully
   - Verify Indexed count matches files processed
   - Verify upsert calls to client
   - Verify document fields are populated correctly

2. **TestIndexFilesWithInvalidUTF8** - Skip files with invalid UTF-8 content
   - Mix valid and invalid UTF-8 files
   - Verify invalid files are skipped (not indexed)
   - Verify warning is logged/captured

3. **TestIndexFilesWithUpsertErrors** - Continue on individual file errors
   - Configure mock to fail on specific upserts
   - Verify other files still get indexed
   - Verify Failed count is accurate

4. **TestIndexFilesEmptyList** - Handle empty file list gracefully
   - Verify result is zero for all counts

5. **TestIndexFilesDocumentBuilding** - Verify document fields are built correctly
   - Check id is hash of absolute path
   - Check file_name is basename
   - Check directory is parent path
   - Check project_root and project_type from projects
   - Check last_modified from file stat
   - Check indexed_at is set to current time

6. **TestIndexFilesProgressLogging** - Verify progress is logged
   - Index multiple files
   - Verify progress callback is called

7. **TestIndexFilesAllFail** - All files fail to index
   - Verify Failed count equals total files
   - Verify Indexed is 0

## Implementation Approach

1. **Types**:
   - `IndexResult` struct with `Indexed`, `Failed`, `Errors` fields
   - Extend `FileInfo` if needed or use existing from sync package

2. **Document Building**:
   - `id`: SHA256 hash of absolute path (first 16 chars hex)
   - `file_name`: `filepath.Base(path)`
   - `directory`: `filepath.Dir(path)`
   - `file_path`: absolute path
   - `project_root` and `project_type`: find matching project by path prefix
   - `last_modified`: from FileInfo
   - `indexed_at`: `time.Now()`
   - `content`: file content (UTF-8 validated)

3. **Error Handling**:
   - Invalid UTF-8: skip file, log warning, don't increment Failed (it's expected behavior)
   - Upsert error: log error, increment Failed, append to Errors, continue to next file

4. **Progress Logging**:
   - Log every N files (e.g., every 100 or 10%)
   - Use callback function for testability

## Files to Create

- `internal/indexer/indexer.go` - Main implementation
- `internal/indexer/indexer_test.go` - Integration tests
