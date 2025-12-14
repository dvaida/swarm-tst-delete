# Plan: Issue #21 - Implement metadata storage (.swarm-indexer-metadata.json)

## Overview

Implement the metadata storage functionality for incremental indexing. This includes reading/writing a `.swarm-indexer-metadata.json` file in each indexed directory, computing content hashes for change detection, and providing a `HasChanged()` method.

## Integration Tests to Write

1. **TestLoad_NonExistentFile** - Load returns empty metadata (not error) when file doesn't exist
2. **TestLoad_ValidFile** - Load correctly parses an existing valid JSON metadata file
3. **TestLoad_CorruptJSON** - Load returns an error when JSON is invalid
4. **TestSave_CreatesFile** - Save creates the metadata file with correct content
5. **TestSave_AtomicWrite** - Save uses atomic write (temp file + rename)
6. **TestSave_OverwritesExisting** - Save overwrites an existing metadata file
7. **TestComputeHash_EmptyDir** - ComputeHash on empty directory returns consistent hash
8. **TestComputeHash_WithFiles** - ComputeHash on directory with files returns hash based on paths+mtimes
9. **TestComputeHash_DetectsChanges** - ComputeHash returns different hash when file mtime changes
10. **TestHasChanged_True** - HasChanged returns true when hash differs
11. **TestHasChanged_False** - HasChanged returns false when hash matches

## Implementation Approach

1. **Metadata Struct**: Define the struct with all required JSON fields
2. **Load()**: Read file, handle non-existent (return empty), handle corrupt (return error)
3. **Save()**: Write to temp file, then rename for atomicity
4. **ComputeHash()**: Walk directory, collect file paths and mtimes, compute SHA256
5. **HasChanged()**: Simple string comparison of stored vs current hash

## Files to Create

- `internal/metadata/metadata.go` - Main implementation
- `internal/metadata/metadata_test.go` - Integration tests
- `go.mod` - Module definition (if not exists)
