# Plan: Issue #19 - Directory Walker with .gitignore Support

## Overview

Implement a file system walker that recursively traverses directories, respects `.gitignore` patterns, detects binary files, and handles symlinks/permissions gracefully.

## Files to Create

- `internal/walker/walker.go` - Main walker implementation
- `internal/walker/binary.go` - Binary file detection
- `internal/walker/walker_test.go` - Integration tests

## Integration Tests to Write

### Walker Tests (`walker_test.go`)

1. **TestWalk_BasicDirectory** - Walk a simple directory structure and verify all files are returned
2. **TestWalk_RespectsGitignore** - Create .gitignore with patterns and verify matching files are skipped
3. **TestWalk_NestedGitignore** - Verify .gitignore patterns at nested levels work correctly
4. **TestWalk_SkipsHiddenDirectories** - Verify hidden directories (starting with `.`) are skipped
5. **TestWalk_AllowsGitDirectory** - Verify `.git` directory is accessible (for project detection)
6. **TestWalk_HandlesSymlinks** - Verify symlinks don't cause infinite loops
7. **TestWalk_HandlesPermissionErrors** - Verify permission errors don't crash the walk
8. **TestWalk_NonExistentRoot** - Verify proper error for non-existent path
9. **TestIsBinary_TextFile** - Verify text files return false
10. **TestIsBinary_BinaryFile** - Verify binary files (with null bytes) return true
11. **TestIsBinary_EmptyFile** - Verify empty files return false
12. **TestIsBinary_NonExistent** - Verify proper error for non-existent file

## Implementation Approach

### Phase 1: Project Setup
- Initialize Go module
- Create directory structure

### Phase 2: Types & Interface
```go
type FileInfo struct {
    Path    string
    Size    int64
    ModTime time.Time
    IsDir   bool
}

func Walk(root string) (<-chan FileInfo, error)
func IsBinary(path string) (bool, error)
```

### Phase 3: Binary Detection
- Open file, read first 8KB
- Check for null bytes (0x00)
- Return true if null bytes found

### Phase 4: Walker Implementation
- Use `filepath.WalkDir` as base
- Parse .gitignore files using a gitignore matching library
- Stack gitignore matchers as we descend into directories
- Skip hidden directories (except .git)
- Track visited inodes to detect symlink loops
- Return FileInfo through channel

### Dependencies
- `github.com/go-git/go-git/v5/plumbing/format/gitignore` for .gitignore parsing (or similar)

## Acceptance Criteria Mapping

| Criteria | Test Coverage |
|----------|---------------|
| Walks directories recursively | TestWalk_BasicDirectory |
| Respects .gitignore at any level | TestWalk_RespectsGitignore, TestWalk_NestedGitignore |
| Correctly identifies binary files | TestIsBinary_* tests |
| Handles symlinks safely | TestWalk_HandlesSymlinks |
| Handles permission errors gracefully | TestWalk_HandlesPermissionErrors |
