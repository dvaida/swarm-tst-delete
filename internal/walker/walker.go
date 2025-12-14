package walker

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
)

// FileInfo represents information about a file to be indexed
type FileInfo struct {
	Path    string
	RelPath string
	Size    int64
	IsDir   bool
}

// Walker traverses directories respecting .gitignore
type Walker struct {
	skipPatterns []string
}

// New creates a new Walker
func New(skipPatterns []string) *Walker {
	return &Walker{
		skipPatterns: skipPatterns,
	}
}

// Walk traverses a directory and returns files to index
func (w *Walker) Walk(ctx context.Context, root string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip hidden directories
		if d.IsDir() && len(d.Name()) > 0 && d.Name()[0] == '.' {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		// Check skip patterns
		for _, pattern := range w.skipPatterns {
			if matched, _ := filepath.Match(pattern, d.Name()); matched {
				return nil
			}
		}

		info, err := d.Info()
		if err != nil {
			return nil // Skip files we can't stat
		}

		relPath, _ := filepath.Rel(root, path)
		files = append(files, FileInfo{
			Path:    path,
			RelPath: relPath,
			Size:    info.Size(),
			IsDir:   false,
		})

		return nil
	})

	return files, err
}

// ReadFile reads the content of a file
func (w *Walker) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
