package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteAndReadMetadata(t *testing.T) {
	root := t.TempDir()

	original := Metadata{
		LastIndexed:    time.Date(2025, 12, 12, 10, 30, 0, 0, time.UTC),
		FilesIndexed:   1523,
		FilesSkipped:   47,
		FilesUnchanged: 890,
		FilesDeleted:   12,
		SkippedFiles: []SkippedFile{
			{Path: "config/.env", Reason: "filename .env"},
			{Path: "src/secrets.go", Reason: "contains password"},
		},
		ProjectsDetected: []DetectedProject{
			{Path: "backend", Type: "go"},
			{Path: "frontend", Type: "node"},
		},
	}

	err := Write(root, original)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	read, err := Read(root)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if !read.LastIndexed.Equal(original.LastIndexed) {
		t.Errorf("LastIndexed mismatch: got %v, want %v", read.LastIndexed, original.LastIndexed)
	}
	if read.FilesIndexed != original.FilesIndexed {
		t.Errorf("FilesIndexed mismatch: got %d, want %d", read.FilesIndexed, original.FilesIndexed)
	}
	if read.FilesSkipped != original.FilesSkipped {
		t.Errorf("FilesSkipped mismatch: got %d, want %d", read.FilesSkipped, original.FilesSkipped)
	}
	if read.FilesUnchanged != original.FilesUnchanged {
		t.Errorf("FilesUnchanged mismatch: got %d, want %d", read.FilesUnchanged, original.FilesUnchanged)
	}
	if read.FilesDeleted != original.FilesDeleted {
		t.Errorf("FilesDeleted mismatch: got %d, want %d", read.FilesDeleted, original.FilesDeleted)
	}
	if len(read.SkippedFiles) != len(original.SkippedFiles) {
		t.Errorf("SkippedFiles length mismatch: got %d, want %d", len(read.SkippedFiles), len(original.SkippedFiles))
	} else {
		for i, sf := range read.SkippedFiles {
			if sf.Path != original.SkippedFiles[i].Path || sf.Reason != original.SkippedFiles[i].Reason {
				t.Errorf("SkippedFiles[%d] mismatch: got %+v, want %+v", i, sf, original.SkippedFiles[i])
			}
		}
	}
	if len(read.ProjectsDetected) != len(original.ProjectsDetected) {
		t.Errorf("ProjectsDetected length mismatch: got %d, want %d", len(read.ProjectsDetected), len(original.ProjectsDetected))
	} else {
		for i, pd := range read.ProjectsDetected {
			if pd.Path != original.ProjectsDetected[i].Path || pd.Type != original.ProjectsDetected[i].Type {
				t.Errorf("ProjectsDetected[%d] mismatch: got %+v, want %+v", i, pd, original.ProjectsDetected[i])
			}
		}
	}
}

func TestMetadataIsPrettyPrinted(t *testing.T) {
	root := t.TempDir()

	meta := Metadata{
		LastIndexed:  time.Date(2025, 12, 12, 10, 30, 0, 0, time.UTC),
		FilesIndexed: 100,
	}

	err := Write(root, meta)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(root, MetadataFilename))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Pretty-printed JSON should contain newlines and indentation
	if !strings.Contains(string(content), "\n") {
		t.Error("JSON should contain newlines (pretty-printed)")
	}
	if !strings.Contains(string(content), "  ") {
		t.Error("JSON should contain indentation (pretty-printed)")
	}
}

func TestTimestampIsRFC3339(t *testing.T) {
	root := t.TempDir()

	meta := Metadata{
		LastIndexed:  time.Date(2025, 12, 12, 10, 30, 0, 0, time.UTC),
		FilesIndexed: 100,
	}

	err := Write(root, meta)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(root, MetadataFilename))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Parse as raw JSON to check timestamp format
	var raw map[string]interface{}
	if err := json.Unmarshal(content, &raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	timestamp, ok := raw["last_indexed"].(string)
	if !ok {
		t.Fatal("last_indexed should be a string")
	}

	// Verify it's RFC3339 format
	if timestamp != "2025-12-12T10:30:00Z" {
		t.Errorf("Timestamp should be RFC3339 format: got %q, want %q", timestamp, "2025-12-12T10:30:00Z")
	}
}

func TestAtomicWriteAndFilePermissions(t *testing.T) {
	root := t.TempDir()

	meta := Metadata{
		LastIndexed:  time.Now().UTC(),
		FilesIndexed: 50,
	}

	err := Write(root, meta)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	filePath := filepath.Join(root, MetadataFilename)
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// Check file permissions are 0644
	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("File permissions should be 0644, got %o", perm)
	}
}

func TestReadNonExistentFile(t *testing.T) {
	root := t.TempDir()

	_, err := Read(root)
	if err == nil {
		t.Error("Read should return error for non-existent file")
	}
}

func TestPathsAreRelative(t *testing.T) {
	root := t.TempDir()

	meta := Metadata{
		LastIndexed: time.Now().UTC(),
		SkippedFiles: []SkippedFile{
			{Path: "config/.env", Reason: "hidden"},
		},
		ProjectsDetected: []DetectedProject{
			{Path: "backend", Type: "go"},
		},
	}

	err := Write(root, meta)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	read, err := Read(root)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Paths should remain relative, not be modified
	if read.SkippedFiles[0].Path != "config/.env" {
		t.Errorf("SkippedFiles path should remain relative: got %q", read.SkippedFiles[0].Path)
	}
	if read.ProjectsDetected[0].Path != "backend" {
		t.Errorf("ProjectsDetected path should remain relative: got %q", read.ProjectsDetected[0].Path)
	}
}

func TestMetadataFilenameConstant(t *testing.T) {
	if MetadataFilename != ".swarm-indexer-metadata.json" {
		t.Errorf("MetadataFilename should be '.swarm-indexer-metadata.json', got %q", MetadataFilename)
	}
}
