package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// MetadataFileName is the name of the metadata file in each indexed directory.
const MetadataFileName = ".swarm-indexer-metadata.json"

// Metadata stores indexing state for a directory.
type Metadata struct {
	LastIndexed  int64             `json:"last_indexed"`
	FileCount    int               `json:"file_count"`
	ContentHash  string            `json:"content_hash"`
	ProjectType  string            `json:"project_type"`
	Languages    []string          `json:"languages"`
	Dependencies map[string]string `json:"dependencies"`
}

// Load reads metadata from the given directory.
// Returns empty metadata if file doesn't exist.
// Returns error if file exists but is corrupt.
func Load(dirPath string) (*Metadata, error) {
	metaPath := filepath.Join(dirPath, MetadataFileName)

	data, err := os.ReadFile(metaPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Metadata{}, nil
		}
		return nil, err
	}

	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return &m, nil
}

// Save writes metadata to the given directory atomically.
func (m *Metadata) Save(dirPath string) error {
	metaPath := filepath.Join(dirPath, MetadataFileName)

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Write to temp file first for atomic save
	tmpPath := metaPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp metadata file: %w", err)
	}

	// Rename for atomic update
	if err := os.Rename(tmpPath, metaPath); err != nil {
		os.Remove(tmpPath) // Clean up temp file on failure
		return fmt.Errorf("failed to rename metadata file: %w", err)
	}

	return nil
}

// ComputeHash computes a hash of file paths and mtimes in the directory.
// This is used for change detection.
func ComputeHash(dirPath string) (string, error) {
	var entries []string

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Skip the metadata file itself
		if d.Name() == MetadataFileName {
			return nil
		}

		// Get relative path for consistent hashing
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		// Include path and mtime in hash input
		entry := fmt.Sprintf("%s:%d", relPath, info.ModTime().UnixNano())
		entries = append(entries, entry)

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk directory: %w", err)
	}

	// Sort for consistent ordering
	sort.Strings(entries)

	// Compute hash
	h := sha256.New()
	for _, entry := range entries {
		h.Write([]byte(entry))
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// HasChanged returns true if the current hash differs from the stored hash.
func (m *Metadata) HasChanged(currentHash string) bool {
	return m.ContentHash != currentHash
}
