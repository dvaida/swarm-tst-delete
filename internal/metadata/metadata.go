// Package metadata provides functionality for writing and reading
// indexing metadata files at registered path roots.
package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// MetadataFilename is the name of the metadata file written to each root.
const MetadataFilename = ".swarm-indexer-metadata.json"

// SkippedFile represents a file that was skipped during indexing.
type SkippedFile struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// DetectedProject represents a project detected during indexing.
type DetectedProject struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

// Metadata contains indexing statistics for a registered path root.
type Metadata struct {
	LastIndexed      time.Time         `json:"last_indexed"`
	FilesIndexed     int               `json:"files_indexed"`
	FilesSkipped     int               `json:"files_skipped"`
	FilesUnchanged   int               `json:"files_unchanged"`
	FilesDeleted     int               `json:"files_deleted"`
	SkippedFiles     []SkippedFile     `json:"skipped_files"`
	ProjectsDetected []DetectedProject `json:"projects_detected"`
}

// Write writes metadata to the specified root directory.
// It uses atomic write (temp file + rename) to avoid corruption.
func Write(root string, meta Metadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(root, MetadataFilename)

	// Create temp file in same directory for atomic rename
	tmpFile, err := os.CreateTemp(root, ".metadata-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Set permissions before rename
	if err := os.Chmod(tmpPath, 0644); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

// Read reads metadata from the specified root directory.
func Read(root string) (*Metadata, error) {
	filePath := filepath.Join(root, MetadataFilename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}
