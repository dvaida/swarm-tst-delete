# Plan: Issue #19 - Directory Walker with .gitignore Support

## Overview

Implement a file system walker that recursively traverses directories, respects `.gitignore` patterns, and detects binary files. The walker returns a channel of `FileInfo` structs for processing.

## Files to Create

1. `internal/walker/walker.go` - Main walker implementation
2. `internal/walker/binary.go` - Binary file detection
3. `internal/walker/walker_test.go` - Integration tests

## Integration Tests

### Test 1: Basic Directory Walk
- Create a temp directory with nested structure
- Verify all files are found recursively
- Verify directories are skipped (only files returned)

### Test 2: .gitignore at Root Level
- Create temp dir with `.gitignore` containing patterns
- Verify ignored files/dirs are skipped
- Verify non-ignored files are returned

### Test 3: Nested .gitignore Files
- Create nested dirs with different `.gitignore` files at each level
- Verify patterns apply only from their level down
- Verify parent patterns still apply to children

### Test 4: Binary File Detection
- Create binary file (with null bytes in first 8KB)
- Create text file (no null bytes)
- Verify `IsBinary()` correctly identifies each

### Test 5: Hidden Directories
- Create hidden directories (starting with `.`)
- Verify they are skipped
- Verify `.git` is also skipped (it's hidden)

### Test 6: Symlink Safety
- Create symlink loop (a -> b -> a)
- Verify walker doesn't loop infinitely
- Verify symlinks to files work correctly

### Test 7: Permission Errors
- Create file with no read permissions
- Verify walker continues gracefully (skips file, doesn't crash)

## Implementation Approach

1. **Walker**: Use `filepath.WalkDir` as base, with custom skip logic
2. **Gitignore**: Use `go-git/go-git/v5/plumbing/format/gitignore` or similar for pattern matching
3. **Binary Detection**: Read first 8KB, check for null bytes
4. **Symlink Handling**: Track visited inodes to detect cycles

## Dependencies

- Standard library (`os`, `path/filepath`, `io`)
- Gitignore pattern matcher (evaluate: `sabhiram/go-gitignore` or similar)
