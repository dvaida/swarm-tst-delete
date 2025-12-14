# Plan for Issue #11: Metadata file writer

## Overview
Implement a metadata package that writes `.swarm-indexer-metadata.json` files containing indexing statistics to registered path roots.

## Requirements Summary
- Create `internal/metadata/metadata.go` package
- `Metadata` struct matching the schema
- `Write(root string, meta Metadata) error` function
- `Read(root string) (*Metadata, error)` function
- JSON pretty-printed with indentation
- Paths relative to root
- Timestamps in RFC3339 format
- Atomic writes (temp file + rename)
- File permissions: 0644

## Integration Tests to Write

1. **TestWriteAndRead** - Write metadata to a temp directory, then read it back and verify all fields match
2. **TestWritePrettyPrinted** - Verify output JSON is indented/pretty-printed
3. **TestTimestampRFC3339Format** - Verify timestamps are in RFC3339 format
4. **TestAtomicWrite** - Verify atomic write behavior (temp file + rename)
5. **TestFilePermissions** - Verify file is created with 0644 permissions
6. **TestReadNonexistent** - Read from path without metadata file returns appropriate error
7. **TestPathsRelativeToRoot** - Verify skipped_files and projects_detected paths are stored relative to root

## Implementation Approach

1. Define the structs:
   - `SkippedFile` with `Path` and `Reason` fields
   - `DetectedProject` with `Path` and `Type` fields
   - `Metadata` with all required fields (timestamps, counts, slices)

2. Implement `Write`:
   - Marshal struct to JSON with `json.MarshalIndent`
   - Write to temp file in same directory
   - Rename temp file to final `.swarm-indexer-metadata.json`
   - Handle errors appropriately

3. Implement `Read`:
   - Read file contents
   - Unmarshal JSON to Metadata struct
   - Return appropriate error if file doesn't exist

## File Structure
```
go.mod
internal/
  metadata/
    metadata.go
    metadata_test.go
```
