package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_NonExistentFile(t *testing.T) {
	// Create a temp directory without metadata file
	tmpDir := t.TempDir()

	// Load should return empty metadata, not an error
	m, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() returned error for non-existent file: %v", err)
	}

	// Should return empty/zero metadata
	if m.LastIndexed != 0 {
		t.Errorf("expected LastIndexed=0, got %d", m.LastIndexed)
	}
	if m.FileCount != 0 {
		t.Errorf("expected FileCount=0, got %d", m.FileCount)
	}
	if m.ContentHash != "" {
		t.Errorf("expected ContentHash='', got %q", m.ContentHash)
	}
	if m.ProjectType != "" {
		t.Errorf("expected ProjectType='', got %q", m.ProjectType)
	}
	if m.Languages != nil {
		t.Errorf("expected Languages=nil, got %v", m.Languages)
	}
	if m.Dependencies != nil {
		t.Errorf("expected Dependencies=nil, got %v", m.Dependencies)
	}
}

func TestLoad_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid metadata file
	meta := Metadata{
		LastIndexed:  1700000000,
		FileCount:    42,
		ContentHash:  "abc123",
		ProjectType:  "go",
		Languages:    []string{"go", "markdown"},
		Dependencies: map[string]string{"cobra": "v1.8.0"},
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("failed to marshal test metadata: %v", err)
	}

	metaPath := filepath.Join(tmpDir, ".swarm-indexer-metadata.json")
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		t.Fatalf("failed to write test metadata file: %v", err)
	}

	// Load should parse correctly
	loaded, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if loaded.LastIndexed != meta.LastIndexed {
		t.Errorf("LastIndexed: expected %d, got %d", meta.LastIndexed, loaded.LastIndexed)
	}
	if loaded.FileCount != meta.FileCount {
		t.Errorf("FileCount: expected %d, got %d", meta.FileCount, loaded.FileCount)
	}
	if loaded.ContentHash != meta.ContentHash {
		t.Errorf("ContentHash: expected %q, got %q", meta.ContentHash, loaded.ContentHash)
	}
	if loaded.ProjectType != meta.ProjectType {
		t.Errorf("ProjectType: expected %q, got %q", meta.ProjectType, loaded.ProjectType)
	}
	if len(loaded.Languages) != len(meta.Languages) {
		t.Errorf("Languages: expected %v, got %v", meta.Languages, loaded.Languages)
	}
	if len(loaded.Dependencies) != len(meta.Dependencies) {
		t.Errorf("Dependencies: expected %v, got %v", meta.Dependencies, loaded.Dependencies)
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a corrupt JSON file
	metaPath := filepath.Join(tmpDir, ".swarm-indexer-metadata.json")
	if err := os.WriteFile(metaPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("failed to write corrupt metadata file: %v", err)
	}

	// Load should return an error
	_, err := Load(tmpDir)
	if err == nil {
		t.Fatal("Load() should return error for corrupt JSON")
	}
}

func TestSave_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()

	meta := &Metadata{
		LastIndexed:  1700000000,
		FileCount:    10,
		ContentHash:  "hash123",
		ProjectType:  "python",
		Languages:    []string{"python"},
		Dependencies: map[string]string{"requests": "2.31.0"},
	}

	if err := meta.Save(tmpDir); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify file exists and has correct content
	metaPath := filepath.Join(tmpDir, ".swarm-indexer-metadata.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("failed to read saved metadata file: %v", err)
	}

	var loaded Metadata
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to parse saved metadata: %v", err)
	}

	if loaded.LastIndexed != meta.LastIndexed {
		t.Errorf("LastIndexed: expected %d, got %d", meta.LastIndexed, loaded.LastIndexed)
	}
	if loaded.FileCount != meta.FileCount {
		t.Errorf("FileCount: expected %d, got %d", meta.FileCount, loaded.FileCount)
	}
	if loaded.ContentHash != meta.ContentHash {
		t.Errorf("ContentHash: expected %q, got %q", meta.ContentHash, loaded.ContentHash)
	}
}

func TestSave_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Save initial metadata
	meta1 := &Metadata{
		LastIndexed: 1000000000,
		FileCount:   5,
		ContentHash: "old_hash",
	}
	if err := meta1.Save(tmpDir); err != nil {
		t.Fatalf("first Save() returned error: %v", err)
	}

	// Save new metadata
	meta2 := &Metadata{
		LastIndexed: 2000000000,
		FileCount:   15,
		ContentHash: "new_hash",
	}
	if err := meta2.Save(tmpDir); err != nil {
		t.Fatalf("second Save() returned error: %v", err)
	}

	// Verify the file has new content
	loaded, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if loaded.LastIndexed != meta2.LastIndexed {
		t.Errorf("LastIndexed: expected %d, got %d", meta2.LastIndexed, loaded.LastIndexed)
	}
	if loaded.FileCount != meta2.FileCount {
		t.Errorf("FileCount: expected %d, got %d", meta2.FileCount, loaded.FileCount)
	}
	if loaded.ContentHash != meta2.ContentHash {
		t.Errorf("ContentHash: expected %q, got %q", meta2.ContentHash, loaded.ContentHash)
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()

	meta := &Metadata{
		LastIndexed: 1700000000,
		FileCount:   10,
		ContentHash: "hash123",
	}

	// Save the metadata
	if err := meta.Save(tmpDir); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify no temp files are left behind
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}

	for _, entry := range entries {
		if entry.Name() != ".swarm-indexer-metadata.json" {
			t.Errorf("unexpected file left behind: %s", entry.Name())
		}
	}

	// Verify exactly one file exists (the metadata file)
	if len(entries) != 1 {
		t.Errorf("expected 1 file, got %d", len(entries))
	}
}

func TestComputeHash_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	hash1, err := ComputeHash(tmpDir)
	if err != nil {
		t.Fatalf("ComputeHash() returned error: %v", err)
	}

	// Hash should be non-empty
	if hash1 == "" {
		t.Error("ComputeHash() returned empty hash for empty directory")
	}

	// Hash should be consistent
	hash2, err := ComputeHash(tmpDir)
	if err != nil {
		t.Fatalf("ComputeHash() returned error on second call: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("ComputeHash() not consistent: %q != %q", hash1, hash2)
	}
}

func TestComputeHash_WithFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some files
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create file1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to create file2: %v", err)
	}

	hash1, err := ComputeHash(tmpDir)
	if err != nil {
		t.Fatalf("ComputeHash() returned error: %v", err)
	}

	if hash1 == "" {
		t.Error("ComputeHash() returned empty hash")
	}

	// Hash should be consistent for same files
	hash2, err := ComputeHash(tmpDir)
	if err != nil {
		t.Fatalf("ComputeHash() returned error on second call: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("ComputeHash() not consistent: %q != %q", hash1, hash2)
	}
}

func TestComputeHash_DetectsChanges(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	filePath := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	hash1, err := ComputeHash(tmpDir)
	if err != nil {
		t.Fatalf("ComputeHash() returned error: %v", err)
	}

	// Wait to ensure mtime will be different
	time.Sleep(10 * time.Millisecond)

	// Modify the file (changing mtime)
	newTime := time.Now().Add(time.Hour)
	if err := os.Chtimes(filePath, newTime, newTime); err != nil {
		t.Fatalf("failed to change file mtime: %v", err)
	}

	hash2, err := ComputeHash(tmpDir)
	if err != nil {
		t.Fatalf("ComputeHash() returned error after modification: %v", err)
	}

	if hash1 == hash2 {
		t.Error("ComputeHash() should return different hash after file modification")
	}
}

func TestComputeHash_IgnoresMetadataFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular file
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	hash1, err := ComputeHash(tmpDir)
	if err != nil {
		t.Fatalf("ComputeHash() returned error: %v", err)
	}

	// Create metadata file - this should not change the hash
	meta := &Metadata{LastIndexed: 123}
	if err := meta.Save(tmpDir); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	hash2, err := ComputeHash(tmpDir)
	if err != nil {
		t.Fatalf("ComputeHash() returned error after adding metadata: %v", err)
	}

	if hash1 != hash2 {
		t.Error("ComputeHash() should ignore .swarm-indexer-metadata.json")
	}
}

func TestHasChanged_True(t *testing.T) {
	meta := &Metadata{
		ContentHash: "old_hash",
	}

	if !meta.HasChanged("new_hash") {
		t.Error("HasChanged() should return true when hashes differ")
	}
}

func TestHasChanged_False(t *testing.T) {
	meta := &Metadata{
		ContentHash: "same_hash",
	}

	if meta.HasChanged("same_hash") {
		t.Error("HasChanged() should return false when hashes match")
	}
}

func TestHasChanged_EmptyStoredHash(t *testing.T) {
	meta := &Metadata{
		ContentHash: "",
	}

	// Empty stored hash means first index, should indicate change needed
	if !meta.HasChanged("any_hash") {
		t.Error("HasChanged() should return true when stored hash is empty")
	}
}
