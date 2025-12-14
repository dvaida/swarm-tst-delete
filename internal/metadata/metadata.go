package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

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

// Metadata contains indexing statistics for a registered path.
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
func Write(root string, meta Metadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	finalPath := filepath.Join(root, MetadataFilename)
	tmpFile, err := os.CreateTemp(root, ".swarm-indexer-metadata-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Chmod(tmpPath, 0644); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

// Read reads metadata from the specified root directory.
func Read(root string) (*Metadata, error) {
	data, err := os.ReadFile(filepath.Join(root, MetadataFilename))
	if err != nil {
		return nil, err
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}
