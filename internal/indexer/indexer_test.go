package indexer

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MockTypesenseClient implements TypesenseClient for testing
type MockTypesenseClient struct {
	upsertCalls  []Document
	upsertErrors map[string]error // key is document ID
}

func NewMockClient() *MockTypesenseClient {
	return &MockTypesenseClient{
		upsertCalls:  []Document{},
		upsertErrors: make(map[string]error),
	}
}

func (m *MockTypesenseClient) UpsertDocument(doc Document) error {
	if err, ok := m.upsertErrors[doc.ID]; ok {
		return err
	}
	m.upsertCalls = append(m.upsertCalls, doc)
	return nil
}

func TestIndexFilesSuccess(t *testing.T) {
	// Create temp files with valid UTF-8 content
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")
	os.WriteFile(file1, []byte("package main"), 0644)
	os.WriteFile(file2, []byte("package util"), 0644)

	stat1, _ := os.Stat(file1)
	stat2, _ := os.Stat(file2)

	files := []FileInfo{
		{Path: file1, LastModified: stat1.ModTime()},
		{Path: file2, LastModified: stat2.ModTime()},
	}
	projects := []Project{{Root: tmpDir, Type: "go"}}
	client := NewMockClient()

	result := IndexFiles(files, projects, client)

	if result.Indexed != 2 {
		t.Errorf("expected Indexed=2, got %d", result.Indexed)
	}
	if result.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", result.Failed)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %v", result.Errors)
	}
	if len(client.upsertCalls) != 2 {
		t.Errorf("expected 2 upsert calls, got %d", len(client.upsertCalls))
	}
}

func TestIndexFilesWithInvalidUTF8(t *testing.T) {
	tmpDir := t.TempDir()

	// Valid UTF-8 file
	validFile := filepath.Join(tmpDir, "valid.go")
	os.WriteFile(validFile, []byte("package main"), 0644)

	// Invalid UTF-8 file (invalid continuation byte)
	invalidFile := filepath.Join(tmpDir, "invalid.bin")
	os.WriteFile(invalidFile, []byte{0xff, 0xfe, 0x00, 0x01}, 0644)

	stat1, _ := os.Stat(validFile)
	stat2, _ := os.Stat(invalidFile)

	files := []FileInfo{
		{Path: validFile, LastModified: stat1.ModTime()},
		{Path: invalidFile, LastModified: stat2.ModTime()},
	}
	projects := []Project{{Root: tmpDir, Type: "go"}}
	client := NewMockClient()

	result := IndexFiles(files, projects, client)

	// Invalid UTF-8 files should be skipped, not counted as failures
	if result.Indexed != 1 {
		t.Errorf("expected Indexed=1, got %d", result.Indexed)
	}
	if result.Skipped != 1 {
		t.Errorf("expected Skipped=1, got %d", result.Skipped)
	}
	if result.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", result.Failed)
	}
	if len(client.upsertCalls) != 1 {
		t.Errorf("expected 1 upsert call, got %d", len(client.upsertCalls))
	}
}

func TestIndexFilesWithUpsertErrors(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")
	file3 := filepath.Join(tmpDir, "file3.go")
	os.WriteFile(file1, []byte("package one"), 0644)
	os.WriteFile(file2, []byte("package two"), 0644)
	os.WriteFile(file3, []byte("package three"), 0644)

	stat1, _ := os.Stat(file1)
	stat2, _ := os.Stat(file2)
	stat3, _ := os.Stat(file3)

	files := []FileInfo{
		{Path: file1, LastModified: stat1.ModTime()},
		{Path: file2, LastModified: stat2.ModTime()},
		{Path: file3, LastModified: stat3.ModTime()},
	}
	projects := []Project{{Root: tmpDir, Type: "go"}}
	client := NewMockClient()

	// Configure file2 to fail
	file2ID := DocumentIDFromPath(file2)
	client.upsertErrors[file2ID] = errors.New("upsert failed")

	result := IndexFiles(files, projects, client)

	// File2 fails, but file1 and file3 should succeed
	if result.Indexed != 2 {
		t.Errorf("expected Indexed=2, got %d", result.Indexed)
	}
	if result.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", result.Failed)
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
	if len(client.upsertCalls) != 2 {
		t.Errorf("expected 2 upsert calls, got %d", len(client.upsertCalls))
	}
}

func TestIndexFilesEmptyList(t *testing.T) {
	client := NewMockClient()
	files := []FileInfo{}
	projects := []Project{}

	result := IndexFiles(files, projects, client)

	if result.Indexed != 0 {
		t.Errorf("expected Indexed=0, got %d", result.Indexed)
	}
	if result.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", result.Failed)
	}
	if result.Skipped != 0 {
		t.Errorf("expected Skipped=0, got %d", result.Skipped)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %v", result.Errors)
	}
	if len(client.upsertCalls) != 0 {
		t.Errorf("expected 0 upsert calls, got %d", len(client.upsertCalls))
	}
}

func TestIndexFilesDocumentBuilding(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	os.MkdirAll(projectDir, 0755)

	filePath := filepath.Join(projectDir, "main.go")
	content := "package main\n\nfunc main() {}"
	os.WriteFile(filePath, []byte(content), 0644)

	stat, _ := os.Stat(filePath)
	absPath, _ := filepath.Abs(filePath)

	files := []FileInfo{
		{Path: absPath, LastModified: stat.ModTime()},
	}
	projects := []Project{{Root: projectDir, Type: "go"}}
	client := NewMockClient()

	beforeIndex := time.Now()
	result := IndexFiles(files, projects, client)
	afterIndex := time.Now()

	if result.Indexed != 1 {
		t.Fatalf("expected Indexed=1, got %d", result.Indexed)
	}
	if len(client.upsertCalls) != 1 {
		t.Fatalf("expected 1 upsert call, got %d", len(client.upsertCalls))
	}

	doc := client.upsertCalls[0]

	// Verify ID is hash of absolute path
	expectedID := DocumentIDFromPath(absPath)
	if doc.ID != expectedID {
		t.Errorf("expected ID=%q, got %q", expectedID, doc.ID)
	}

	// Verify file_name is basename
	if doc.FileName != "main.go" {
		t.Errorf("expected FileName=main.go, got %q", doc.FileName)
	}

	// Verify directory is parent path
	if doc.Directory != projectDir {
		t.Errorf("expected Directory=%q, got %q", projectDir, doc.Directory)
	}

	// Verify file_path is absolute path
	if doc.FilePath != absPath {
		t.Errorf("expected FilePath=%q, got %q", absPath, doc.FilePath)
	}

	// Verify project_root and project_type
	if doc.ProjectRoot != projectDir {
		t.Errorf("expected ProjectRoot=%q, got %q", projectDir, doc.ProjectRoot)
	}
	if doc.ProjectType != "go" {
		t.Errorf("expected ProjectType=go, got %q", doc.ProjectType)
	}

	// Verify last_modified matches file stat
	if !doc.LastModified.Equal(stat.ModTime()) {
		t.Errorf("expected LastModified=%v, got %v", stat.ModTime(), doc.LastModified)
	}

	// Verify indexed_at is set to approximately current time
	if doc.IndexedAt.Before(beforeIndex) || doc.IndexedAt.After(afterIndex) {
		t.Errorf("IndexedAt %v should be between %v and %v", doc.IndexedAt, beforeIndex, afterIndex)
	}

	// Verify content
	if doc.Content != content {
		t.Errorf("expected Content=%q, got %q", content, doc.Content)
	}
}

func TestIndexFilesProjectMatching(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two project directories
	goProject := filepath.Join(tmpDir, "goproject")
	pyProject := filepath.Join(tmpDir, "pyproject")
	os.MkdirAll(goProject, 0755)
	os.MkdirAll(pyProject, 0755)

	goFile := filepath.Join(goProject, "main.go")
	pyFile := filepath.Join(pyProject, "main.py")
	os.WriteFile(goFile, []byte("package main"), 0644)
	os.WriteFile(pyFile, []byte("print('hello')"), 0644)

	stat1, _ := os.Stat(goFile)
	stat2, _ := os.Stat(pyFile)

	files := []FileInfo{
		{Path: goFile, LastModified: stat1.ModTime()},
		{Path: pyFile, LastModified: stat2.ModTime()},
	}
	projects := []Project{
		{Root: goProject, Type: "go"},
		{Root: pyProject, Type: "python"},
	}
	client := NewMockClient()

	result := IndexFiles(files, projects, client)

	if result.Indexed != 2 {
		t.Fatalf("expected Indexed=2, got %d", result.Indexed)
	}

	// Verify each file got the correct project
	for _, doc := range client.upsertCalls {
		if strings.Contains(doc.FilePath, "goproject") {
			if doc.ProjectType != "go" {
				t.Errorf("go file should have ProjectType=go, got %q", doc.ProjectType)
			}
			if doc.ProjectRoot != goProject {
				t.Errorf("go file should have ProjectRoot=%q, got %q", goProject, doc.ProjectRoot)
			}
		} else if strings.Contains(doc.FilePath, "pyproject") {
			if doc.ProjectType != "python" {
				t.Errorf("py file should have ProjectType=python, got %q", doc.ProjectType)
			}
			if doc.ProjectRoot != pyProject {
				t.Errorf("py file should have ProjectRoot=%q, got %q", pyProject, doc.ProjectRoot)
			}
		}
	}
}

func TestIndexFilesAllFail(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")
	os.WriteFile(file1, []byte("package one"), 0644)
	os.WriteFile(file2, []byte("package two"), 0644)

	stat1, _ := os.Stat(file1)
	stat2, _ := os.Stat(file2)

	files := []FileInfo{
		{Path: file1, LastModified: stat1.ModTime()},
		{Path: file2, LastModified: stat2.ModTime()},
	}
	projects := []Project{{Root: tmpDir, Type: "go"}}
	client := NewMockClient()

	// Configure all files to fail
	client.upsertErrors[DocumentIDFromPath(file1)] = errors.New("failed 1")
	client.upsertErrors[DocumentIDFromPath(file2)] = errors.New("failed 2")

	result := IndexFiles(files, projects, client)

	if result.Indexed != 0 {
		t.Errorf("expected Indexed=0, got %d", result.Indexed)
	}
	if result.Failed != 2 {
		t.Errorf("expected Failed=2, got %d", result.Failed)
	}
	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
}

func TestIndexFilesFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "nonexistent.go")

	files := []FileInfo{
		{Path: nonExistent, LastModified: time.Now()},
	}
	projects := []Project{{Root: tmpDir, Type: "go"}}
	client := NewMockClient()

	result := IndexFiles(files, projects, client)

	// File not found should be a failure
	if result.Indexed != 0 {
		t.Errorf("expected Indexed=0, got %d", result.Indexed)
	}
	if result.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", result.Failed)
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestDocumentIDFromPath(t *testing.T) {
	// Document IDs should be consistent
	path := "/project/src/main.go"
	id1 := DocumentIDFromPath(path)
	id2 := DocumentIDFromPath(path)

	if id1 != id2 {
		t.Errorf("same path should produce same ID: %q vs %q", id1, id2)
	}

	if id1 == "" {
		t.Error("document ID should not be empty")
	}

	// Should be a hash (16 hex chars)
	if len(id1) != 16 {
		t.Errorf("expected ID length 16, got %d", len(id1))
	}

	// Verify it's valid hex
	_, err := hex.DecodeString(id1)
	if err != nil {
		t.Errorf("ID should be valid hex: %v", err)
	}

	// Different paths should produce different IDs
	id3 := DocumentIDFromPath("/project/src/other.go")
	if id1 == id3 {
		t.Error("different paths should produce different IDs")
	}

	// Verify it's based on SHA256
	hash := sha256.Sum256([]byte(path))
	expected := hex.EncodeToString(hash[:])[:16]
	if id1 != expected {
		t.Errorf("expected ID=%q, got %q", expected, id1)
	}
}

func TestIndexFilesNoMatchingProject(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "orphan.go")
	os.WriteFile(file, []byte("package orphan"), 0644)

	stat, _ := os.Stat(file)

	files := []FileInfo{
		{Path: file, LastModified: stat.ModTime()},
	}
	// Empty projects - file has no matching project
	projects := []Project{}
	client := NewMockClient()

	result := IndexFiles(files, projects, client)

	if result.Indexed != 1 {
		t.Fatalf("expected Indexed=1, got %d", result.Indexed)
	}

	doc := client.upsertCalls[0]

	// File without matching project should have empty project fields
	if doc.ProjectRoot != "" {
		t.Errorf("expected empty ProjectRoot, got %q", doc.ProjectRoot)
	}
	if doc.ProjectType != "" {
		t.Errorf("expected empty ProjectType, got %q", doc.ProjectType)
	}
}
