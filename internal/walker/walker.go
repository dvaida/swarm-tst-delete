package walker

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// FileInfo contains metadata about a file discovered during walking.
type FileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// Walk recursively traverses the directory tree starting at root,
// respecting .gitignore patterns and skipping hidden directories (except .git).
// It returns a channel of FileInfo for each file discovered.
func Walk(root string) (<-chan FileInfo, error) {
	// Verify root exists
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrNotExist
	}

	ch := make(chan FileInfo)

	go func() {
		defer close(ch)
		walkDir(root, root, nil, make(map[uint64]bool), ch)
	}()

	return ch, nil
}

// walkDir recursively walks a directory, building up gitignore patterns.
func walkDir(root, dir string, parentPatterns []gitignore.Pattern, visited map[uint64]bool, ch chan<- FileInfo) {
	// Check for symlink loop by tracking inodes
	if stat, err := os.Stat(dir); err == nil {
		if sysStat, ok := stat.Sys().(*syscall.Stat_t); ok {
			if visited[sysStat.Ino] {
				return // Already visited this inode
			}
			visited[sysStat.Ino] = true
		}
	}

	// Build patterns for this directory
	patterns := buildPatterns(dir, parentPatterns)

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		// Skip directories we can't read (permission errors)
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(dir, name)
		relPath, _ := filepath.Rel(root, path)

		// Skip hidden directories (but allow .git)
		if entry.IsDir() && strings.HasPrefix(name, ".") && name != ".git" {
			continue
		}

		// Skip hidden files that start with . (except .gitignore which we already processed)
		if !entry.IsDir() && strings.HasPrefix(name, ".") {
			continue
		}

		// Check if ignored by gitignore
		if isIgnored(patterns, relPath, entry.IsDir()) {
			continue
		}

		if entry.IsDir() {
			walkDir(root, path, patterns, visited, ch)
		} else {
			info, err := entry.Info()
			if err != nil {
				continue // Skip files we can't stat
			}
			ch <- FileInfo{
				Path:    path,
				Size:    info.Size(),
				ModTime: info.ModTime(),
				IsDir:   false,
			}
		}
	}
}

// buildPatterns creates gitignore patterns for the given directory,
// inheriting patterns from parent.
func buildPatterns(dir string, parentPatterns []gitignore.Pattern) []gitignore.Pattern {
	patterns := parentPatterns

	// Read .gitignore if it exists
	gitignorePath := filepath.Join(dir, ".gitignore")
	if f, err := os.Open(gitignorePath); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			patterns = append(patterns, gitignore.ParsePattern(line, nil))
		}
	}

	return patterns
}

// isIgnored checks if a path should be ignored.
func isIgnored(patterns []gitignore.Pattern, relPath string, isDir bool) bool {
	if len(patterns) == 0 {
		return false
	}

	// Convert path to slice format expected by gitignore
	parts := strings.Split(relPath, string(filepath.Separator))

	// Check patterns in reverse order (later patterns override earlier ones)
	for i := len(patterns) - 1; i >= 0; i-- {
		result := patterns[i].Match(parts, isDir)
		if result == gitignore.Exclude {
			return true
		}
		if result == gitignore.Include {
			return false
		}
	}

	return false
}
