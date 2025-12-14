package walker

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileInfo contains information about a file discovered during walking.
type FileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// Walk recursively traverses the directory tree starting at root,
// respecting .gitignore files at each level and skipping hidden directories.
// It returns a channel that yields FileInfo for each discovered file.
func Walk(root string) (<-chan FileInfo, error) {
	// Verify root exists
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrInvalid
	}

	ch := make(chan FileInfo)

	go func() {
		defer close(ch)
		visited := make(map[uint64]bool) // Track visited inodes to prevent symlink loops
		var patterns []gitignorePattern
		walkDir(root, root, patterns, visited, ch)
	}()

	return ch, nil
}

type gitignorePattern struct {
	pattern string
	negated bool
	dirOnly bool
	baseDir string
}

func walkDir(root, dir string, parentPatterns []gitignorePattern, visited map[uint64]bool, ch chan<- FileInfo) {
	// Load .gitignore for this directory
	patterns := append([]gitignorePattern{}, parentPatterns...)
	gitignorePath := filepath.Join(dir, ".gitignore")
	if newPatterns, err := loadGitignore(gitignorePath, dir); err == nil {
		patterns = append(patterns, newPatterns...)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return // Skip directories we can't read
	}

	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(dir, name)

		// Skip hidden files/directories (starting with .)
		if strings.HasPrefix(name, ".") {
			continue
		}

		// Check gitignore patterns
		relPath, _ := filepath.Rel(root, path)
		if isIgnored(relPath, entry.IsDir(), patterns) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		// Handle symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			// Resolve symlink to check if it's a directory and prevent loops
			realPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				continue // Skip broken symlinks
			}

			realInfo, err := os.Stat(realPath)
			if err != nil {
				continue
			}

			// Get inode to detect loops
			if inode := getInode(realInfo); inode != 0 {
				if visited[inode] {
					continue // Already visited this inode
				}
				visited[inode] = true
			}

			if realInfo.IsDir() {
				// Don't follow symlinks to directories (could cause loops)
				continue
			}

			// For symlinks to files, emit the symlink path but with real file info
			ch <- FileInfo{
				Path:    path,
				Size:    realInfo.Size(),
				ModTime: realInfo.ModTime(),
				IsDir:   false,
			}
			continue
		}

		if entry.IsDir() {
			// Track inode for directories too
			if inode := getInode(info); inode != 0 {
				if visited[inode] {
					continue
				}
				visited[inode] = true
			}
			walkDir(root, path, patterns, visited, ch)
		} else {
			ch <- FileInfo{
				Path:    path,
				Size:    info.Size(),
				ModTime: info.ModTime(),
				IsDir:   false,
			}
		}
	}
}

func loadGitignore(path, baseDir string) ([]gitignorePattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var patterns []gitignorePattern
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pattern := gitignorePattern{baseDir: baseDir}

		// Handle negation
		if strings.HasPrefix(line, "!") {
			pattern.negated = true
			line = line[1:]
		}

		// Handle directory-only patterns
		if strings.HasSuffix(line, "/") {
			pattern.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}

		pattern.pattern = line
		patterns = append(patterns, pattern)
	}

	return patterns, scanner.Err()
}

func isIgnored(relPath string, isDir bool, patterns []gitignorePattern) bool {
	ignored := false

	for _, p := range patterns {
		if p.dirOnly && !isDir {
			continue
		}

		if matchPattern(relPath, p) {
			ignored = !p.negated
		}
	}

	return ignored
}

func matchPattern(relPath string, p gitignorePattern) bool {
	// Get the path relative to where the gitignore was found
	name := filepath.Base(relPath)

	// Simple matching: if pattern doesn't contain /, match against basename
	if !strings.Contains(p.pattern, "/") {
		matched, _ := filepath.Match(p.pattern, name)
		return matched
	}

	// For patterns with /, compute path relative to gitignore location
	// and match against the full relative path
	matched, _ := filepath.Match(p.pattern, relPath)
	return matched
}
