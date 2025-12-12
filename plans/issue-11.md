# Plan: Issue #11 - Metadata file writer

## Overview
Implement a metadata package that writes `.swarm-indexer-metadata.json` files at registered path roots with indexing statistics. The package needs to support both writing and reading metadata files with proper atomic writes.

## Deliverables
- `internal/metadata/metadata.go` - Package implementation
- `internal/metadata/metadata_test.go` - Integration tests

## Data Structures
```go
type SkippedFile struct {
    Path   string `json:"path"`
    Reason string `json:"reason"`
}

type DetectedProject struct {
    Path string `json:"path"`
    Type string `json:"type"`
}

type Metadata struct {
    LastIndexed      time.Time         `json:"last_indexed"`
    FilesIndexed     int               `json:"files_indexed"`
    FilesSkipped     int               `json:"files_skipped"`
    FilesUnchanged   int               `json:"files_unchanged"`
    FilesDeleted     int               `json:"files_deleted"`
    SkippedFiles     []SkippedFile     `json:"skipped_files"`
    ProjectsDetected []DetectedProject `json:"projects_detected"`
}
```

## Integration Tests to Write

1. **TestWriteAndReadMetadata** - Write metadata, then read it back and verify all fields match
2. **TestMetadataIsPrettyPrinted** - Write metadata and verify JSON has indentation
3. **TestTimestampIsRFC3339** - Write metadata and verify timestamp format
4. **TestAtomicWriteAndFilePermissions** - Verify file permissions are 0644 and atomic write works
5. **TestReadNonExistentFile** - Read from non-existent path returns error
6. **TestPathsAreRelative** - Verify paths in skipped_files and projects_detected are stored as provided (relative)
7. **TestMetadataFilenameConstant** - Verify the constant is `.swarm-indexer-metadata.json`

## Implementation Approach

1. Define constants:
   - `MetadataFilename = ".swarm-indexer-metadata.json"`

2. Define types:
   - `SkippedFile` struct
   - `DetectedProject` struct
   - `Metadata` struct

3. Implement `Write(root string, meta Metadata) error`:
   - Marshal metadata to JSON with `json.MarshalIndent` (2-space indent)
   - Write to temporary file in same directory
   - Set permissions to 0644
   - Rename temp file to final filename (atomic operation)

4. Implement `Read(root string) (*Metadata, error)`:
   - Construct full path from root + MetadataFilename
   - Read file contents
   - Unmarshal JSON to Metadata struct
   - Return pointer to metadata

## Technical Notes
- Use `json.MarshalIndent(data, "", "  ")` for pretty-printing
- Use `os.CreateTemp` in same directory for atomic write
- Use `os.Rename` for atomic move
- File permissions: 0644
- Timestamps use `time.Time` which serializes to RFC3339 automatically
