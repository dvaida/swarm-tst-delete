package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()

	meta := Metadata{
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

	err := Write(tmpDir, meta)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	readMeta, err := Read(tmpDir)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if !readMeta.LastIndexed.Equal(meta.LastIndexed) {
		t.Errorf("LastIndexed mismatch: got %v, want %v", readMeta.LastIndexed, meta.LastIndexed)
	}
	if readMeta.FilesIndexed != meta.FilesIndexed {
		t.Errorf("FilesIndexed mismatch: got %d, want %d", readMeta.FilesIndexed, meta.FilesIndexed)
	}
	if readMeta.FilesSkipped != meta.FilesSkipped {
		t.Errorf("FilesSkipped mismatch: got %d, want %d", readMeta.FilesSkipped, meta.FilesSkipped)
	}
	if readMeta.FilesUnchanged != meta.FilesUnchanged {
		t.Errorf("FilesUnchanged mismatch: got %d, want %d", readMeta.FilesUnchanged, meta.FilesUnchanged)
	}
	if readMeta.FilesDeleted != meta.FilesDeleted {
		t.Errorf("FilesDeleted mismatch: got %d, want %d", readMeta.FilesDeleted, meta.FilesDeleted)
	}
	if len(readMeta.SkippedFiles) != len(meta.SkippedFiles) {
		t.Errorf("SkippedFiles length mismatch: got %d, want %d", len(readMeta.SkippedFiles), len(meta.SkippedFiles))
	} else {
		for i, sf := range readMeta.SkippedFiles {
			if sf.Path != meta.SkippedFiles[i].Path || sf.Reason != meta.SkippedFiles[i].Reason {
				t.Errorf("SkippedFiles[%d] mismatch: got %+v, want %+v", i, sf, meta.SkippedFiles[i])
			}
		}
	}
	if len(readMeta.ProjectsDetected) != len(meta.ProjectsDetected) {
		t.Errorf("ProjectsDetected length mismatch: got %d, want %d", len(readMeta.ProjectsDetected), len(meta.ProjectsDetected))
	} else {
		for i, pd := range readMeta.ProjectsDetected {
			if pd.Path != meta.ProjectsDetected[i].Path || pd.Type != meta.ProjectsDetected[i].Type {
				t.Errorf("ProjectsDetected[%d] mismatch: got %+v, want %+v", i, pd, meta.ProjectsDetected[i])
			}
		}
	}
}

func TestWritePrettyPrinted(t *testing.T) {
	tmpDir := t.TempDir()

	meta := Metadata{
		LastIndexed:      time.Date(2025, 12, 12, 10, 30, 0, 0, time.UTC),
		FilesIndexed:     100,
		FilesSkipped:     10,
		FilesUnchanged:   50,
		FilesDeleted:     5,
		SkippedFiles:     []SkippedFile{},
		ProjectsDetected: []DetectedProject{},
	}

	err := Write(tmpDir, meta)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, MetadataFilename))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Pretty-printed JSON should contain newlines and indentation
	if !strings.Contains(string(content), "\n") {
		t.Error("JSON is not pretty-printed: no newlines found")
	}
	if !strings.Contains(string(content), "  ") {
		t.Error("JSON is not pretty-printed: no indentation found")
	}
}

func TestTimestampRFC3339Format(t *testing.T) {
	tmpDir := t.TempDir()

	meta := Metadata{
		LastIndexed:      time.Date(2025, 12, 12, 10, 30, 0, 0, time.UTC),
		FilesIndexed:     100,
		FilesSkipped:     10,
		FilesUnchanged:   50,
		FilesDeleted:     5,
		SkippedFiles:     []SkippedFile{},
		ProjectsDetected: []DetectedProject{},
	}

	err := Write(tmpDir, meta)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, MetadataFilename))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Parse as generic JSON to check timestamp format
	var raw map[string]interface{}
	if err := json.Unmarshal(content, &raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	lastIndexed, ok := raw["last_indexed"].(string)
	if !ok {
		t.Fatal("last_indexed is not a string")
	}

	// Verify it parses as RFC3339
	_, err = time.Parse(time.RFC3339, lastIndexed)
	if err != nil {
		t.Errorf("last_indexed is not RFC3339 format: %v", err)
	}

	// Check expected value
	if lastIndexed != "2025-12-12T10:30:00Z" {
		t.Errorf("last_indexed mismatch: got %s, want 2025-12-12T10:30:00Z", lastIndexed)
	}
}

func TestAtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	metaPath := filepath.Join(tmpDir, MetadataFilename)

	// Write initial metadata
	meta1 := Metadata{
		LastIndexed:      time.Date(2025, 12, 12, 10, 0, 0, 0, time.UTC),
		FilesIndexed:     100,
		SkippedFiles:     []SkippedFile{},
		ProjectsDetected: []DetectedProject{},
	}
	if err := Write(tmpDir, meta1); err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Fatal("Metadata file was not created")
	}

	// Write again (overwrite)
	meta2 := Metadata{
		LastIndexed:      time.Date(2025, 12, 12, 11, 0, 0, 0, time.UTC),
		FilesIndexed:     200,
		SkippedFiles:     []SkippedFile{},
		ProjectsDetected: []DetectedProject{},
	}
	if err := Write(tmpDir, meta2); err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	// Read back and verify it's the second write
	readMeta, err := Read(tmpDir)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if readMeta.FilesIndexed != 200 {
		t.Errorf("Expected FilesIndexed=200, got %d", readMeta.FilesIndexed)
	}

	// Verify no temp files left behind
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".swarm-indexer-metadata") && entry.Name() != MetadataFilename {
			t.Errorf("Temp file left behind: %s", entry.Name())
		}
	}
}

func TestFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	meta := Metadata{
		LastIndexed:      time.Date(2025, 12, 12, 10, 30, 0, 0, time.UTC),
		FilesIndexed:     100,
		SkippedFiles:     []SkippedFile{},
		ProjectsDetected: []DetectedProject{},
	}

	err := Write(tmpDir, meta)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	info, err := os.Stat(filepath.Join(tmpDir, MetadataFilename))
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("File permissions mismatch: got %o, want 0644", perm)
	}
}

func TestReadNonexistent(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Read(tmpDir)
	if err == nil {
		t.Error("Expected error when reading nonexistent file, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.IsNotExist error, got: %v", err)
	}
}

func TestPathsRelativeToRoot(t *testing.T) {
	tmpDir := t.TempDir()

	// Paths should already be relative when passed in
	meta := Metadata{
		LastIndexed:  time.Date(2025, 12, 12, 10, 30, 0, 0, time.UTC),
		FilesIndexed: 100,
		SkippedFiles: []SkippedFile{
			{Path: "config/.env", Reason: "filename .env"},
			{Path: "deep/nested/path/secret.txt", Reason: "contains secret"},
		},
		ProjectsDetected: []DetectedProject{
			{Path: "backend", Type: "go"},
			{Path: "services/api", Type: "node"},
		},
	}

	err := Write(tmpDir, meta)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, MetadataFilename))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Verify paths in JSON are relative (not absolute)
	contentStr := string(content)
	if strings.Contains(contentStr, tmpDir) {
		t.Error("JSON contains absolute paths, expected relative paths")
	}
	if !strings.Contains(contentStr, "config/.env") {
		t.Error("Expected relative path 'config/.env' in JSON")
	}
	if !strings.Contains(contentStr, "deep/nested/path/secret.txt") {
		t.Error("Expected relative path 'deep/nested/path/secret.txt' in JSON")
	}
	if !strings.Contains(contentStr, "services/api") {
		t.Error("Expected relative path 'services/api' in JSON")
	}
}
