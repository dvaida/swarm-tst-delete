# Plan for Issue #19: Implement Directory Walker with .gitignore Support

## Overview
Create a file system walker that recursively traverses directories, respects `.gitignore` patterns, detects binary files, and handles symlinks safely. This is a foundational component for the swarm-indexer tool.

## Files to Create
- `internal/walker/walker.go` - Main walking logic and FileInfo type
- `internal/walker/binary.go` - Binary file detection
- `internal/walker/walker_test.go` - Integration tests

## Integration Tests to Write

### 1. TestWalk_BasicDirectory
- Creates a temp directory with nested files
- Verifies all files are returned via channel
- Verifies FileInfo fields are populated correctly

### 2. TestWalk_RespectsGitignore
- Creates temp directory with `.gitignore` file
- Creates files that match and don't match patterns
- Verifies ignored files are not returned

### 3. TestWalk_NestedGitignore
- Creates nested directories each with `.gitignore`
- Verifies patterns are applied at correct levels

### 4. TestWalk_SkipsHiddenDirectories
- Creates hidden directories (starting with `.`)
- Verifies they are skipped (except `.git`)

### 5. TestWalk_HandlesSymlinks
- Creates symlinks (including circular)
- Verifies no infinite loops occur
- Verifies symlinks to files are handled

### 6. TestWalk_PermissionErrors
- Creates directory with restricted permissions
- Verifies walker continues gracefully

### 7. TestIsBinary_TextFile
- Creates text files with various content
- Verifies IsBinary returns false

### 8. TestIsBinary_BinaryFile
- Creates files with null bytes in first 8KB
- Verifies IsBinary returns true

### 9. TestIsBinary_EmptyFile
- Empty file should not be binary

### 10. TestWalk_NonExistentPath
- Verifies appropriate error for non-existent path

## Implementation Approach

### walker.go
1. Define `FileInfo` struct with Path, Size, ModTime, IsDir
2. Implement `Walk(root string) (<-chan FileInfo, error)`:
   - Validate root exists
   - Create output channel
   - Launch goroutine to walk directories
   - Use `filepath.WalkDir` or manual traversal
   - Load and stack `.gitignore` patterns per directory
   - Skip hidden directories (except `.git`)
   - Track visited inodes to prevent symlink loops
   - Close channel when done

### binary.go
1. Implement `IsBinary(path string) (bool, error)`:
   - Open file, read first 8KB
   - Check for null bytes (0x00)
   - Return true if null bytes found

## Dependencies
- Use `github.com/go-git/go-git/v5/plumbing/format/gitignore` for gitignore parsing
- Standard library for file operations

## Acceptance Criteria Mapping
- [x] Walks directories recursively → TestWalk_BasicDirectory
- [x] Respects .gitignore at any level → TestWalk_RespectsGitignore, TestWalk_NestedGitignore
- [x] Correctly identifies binary files → TestIsBinary_*
- [x] Handles symlinks safely → TestWalk_HandlesSymlinks
- [x] Handles permission errors gracefully → TestWalk_PermissionErrors
