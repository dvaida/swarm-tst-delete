// Package walker provides directory traversal with .gitignore support.
package walker

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	ignore "github.com/sabhiram/go-gitignore"
)

// FileInfo contains metadata about a discovered file.
type FileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// Walk recursively traverses the directory tree starting at root,
// respecting .gitignore patterns and skipping hidden directories.
// It returns a channel of FileInfo for each discovered file.
func Walk(root string) (<-chan FileInfo, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		ch := make(chan FileInfo)
		close(ch)
		return ch, nil
	}

	ch := make(chan FileInfo)

	go func() {
		defer close(ch)

		// Track visited directories by their real path to detect symlink loops
		visited := make(map[string]bool)

		// Stack of gitignore matchers (accumulated through directory descent)
		var gitignores []*ignore.GitIgnore

		var walkFn func(dir string) error
		walkFn = func(dir string) error {
			// Check for symlink loop
			realPath, err := filepath.EvalSymlinks(dir)
			if err != nil {
				// Skip if we can't resolve symlink
				return nil
			}
			if visited[realPath] {
				// Already visited, skip to avoid infinite loop
				return nil
			}
			visited[realPath] = true

			// Load .gitignore if present in this directory
			gitignorePath := filepath.Join(dir, ".gitignore")
			var localIgnore *ignore.GitIgnore
			if gi, err := ignore.CompileIgnoreFile(gitignorePath); err == nil {
				localIgnore = gi
				gitignores = append(gitignores, gi)
			}

			// Read directory entries
			entries, err := os.ReadDir(dir)
			if err != nil {
				// Permission error or other issues - skip this directory
				return nil
			}

			for _, entry := range entries {
				name := entry.Name()
				fullPath := filepath.Join(dir, name)

				// Skip hidden directories (except at root for .gitignore itself)
				if entry.IsDir() && strings.HasPrefix(name, ".") {
					continue
				}

				// Check gitignore patterns
				relPath, _ := filepath.Rel(absRoot, fullPath)
				if isIgnored(relPath, entry.IsDir(), gitignores) {
					continue
				}

				if entry.IsDir() {
					// Recurse into subdirectory
					if err := walkFn(fullPath); err != nil {
						return err
					}
				} else {
					// Get file info
					info, err := entry.Info()
					if err != nil {
						continue // Skip files we can't stat
					}

					ch <- FileInfo{
						Path:    fullPath,
						Size:    info.Size(),
						ModTime: info.ModTime(),
						IsDir:   false,
					}
				}
			}

			// Pop local gitignore from stack when leaving directory
			if localIgnore != nil {
				gitignores = gitignores[:len(gitignores)-1]
			}

			return nil
		}

		_ = walkFn(absRoot)
	}()

	return ch, nil
}

// isIgnored checks if a path matches any of the gitignore patterns
func isIgnored(relPath string, isDir bool, gitignores []*ignore.GitIgnore) bool {
	// Normalize path for matching
	checkPath := relPath
	if isDir {
		checkPath = relPath + "/"
	}

	for _, gi := range gitignores {
		if gi.MatchesPath(checkPath) {
			return true
		}
		// Also check without trailing slash for directories
		if isDir && gi.MatchesPath(relPath) {
			return true
		}
	}
	return false
}
