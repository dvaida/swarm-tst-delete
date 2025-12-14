package status_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dvaida/swarm-indexer/internal/indexer"
	"github.com/dvaida/swarm-indexer/internal/metadata"
	"github.com/dvaida/swarm-indexer/internal/status"
)

func TestStatusCommand_ShowsPathMetadata(t *testing.T) {
	// Setup: Create a temp directory with metadata
	tmpDir := t.TempDir()
	meta := &metadata.Metadata{
		LastIndexed:  time.Date(2024, 1, 15, 14, 30, 22, 0, time.UTC).Unix(),
		FileCount:    1234,
		ContentHash:  "abc123",
		ProjectType:  "node",
		Languages:    []string{"ts", "js", "json"},
		Dependencies: map[string]string{},
	}
	writeMetadata(t, tmpDir, meta)

	// Create mock Typesense client that returns stats
	mockClient := &indexer.MockClient{
		Stats: &indexer.CollectionStats{
			DocumentCount: 45678,
			CollectionName: "swarm-index",
		},
	}

	var buf bytes.Buffer
	err := status.Run([]string{tmpDir}, mockClient, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check path is displayed
	if !strings.Contains(output, tmpDir) {
		t.Errorf("output should contain path %s, got: %s", tmpDir, output)
	}

	// Check type is displayed
	if !strings.Contains(output, "node") {
		t.Errorf("output should contain project type 'node', got: %s", output)
	}

	// Check file count is displayed
	if !strings.Contains(output, "1,234") {
		t.Errorf("output should contain formatted file count '1,234', got: %s", output)
	}

	// Check languages are displayed
	if !strings.Contains(output, "ts") || !strings.Contains(output, "js") || !strings.Contains(output, "json") {
		t.Errorf("output should contain languages, got: %s", output)
	}

	// Check last indexed is displayed
	if !strings.Contains(output, "2024-01-15") {
		t.Errorf("output should contain last indexed date, got: %s", output)
	}
}

func TestStatusCommand_DetectsChanges(t *testing.T) {
	// Setup: Create a temp directory with metadata and a file that changes the hash
	tmpDir := t.TempDir()

	// Write a file that will be hashed
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("original content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Compute the current hash
	currentHash, err := metadata.ComputeContentHash(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Write metadata with a different hash (simulating changes)
	meta := &metadata.Metadata{
		LastIndexed:  time.Now().Unix(),
		FileCount:    1,
		ContentHash:  "different-hash-than-current",
		ProjectType:  "unknown",
		Languages:    []string{},
		Dependencies: map[string]string{},
	}
	writeMetadata(t, tmpDir, meta)

	// Verify hashes are different
	if currentHash == meta.ContentHash {
		t.Fatal("test setup error: hashes should be different")
	}

	mockClient := &indexer.MockClient{
		Stats: &indexer.CollectionStats{
			DocumentCount: 100,
			CollectionName: "swarm-index",
		},
	}

	var buf bytes.Buffer
	err = status.Run([]string{tmpDir}, mockClient, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check that "Changes detected" or similar is shown
	if !strings.Contains(output, "Changes detected") && !strings.Contains(output, "re-index needed") {
		t.Errorf("output should indicate changes detected, got: %s", output)
	}
}

func TestStatusCommand_ShowsUpToDate(t *testing.T) {
	// Setup: Create a temp directory with metadata where hash matches
	tmpDir := t.TempDir()

	// Write a file that will be hashed
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Compute the current hash
	currentHash, err := metadata.ComputeContentHash(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Write metadata with matching hash
	meta := &metadata.Metadata{
		LastIndexed:  time.Now().Unix(),
		FileCount:    1,
		ContentHash:  currentHash,
		ProjectType:  "unknown",
		Languages:    []string{},
		Dependencies: map[string]string{},
	}
	writeMetadata(t, tmpDir, meta)

	mockClient := &indexer.MockClient{
		Stats: &indexer.CollectionStats{
			DocumentCount: 100,
			CollectionName: "swarm-index",
		},
	}

	var buf bytes.Buffer
	err = status.Run([]string{tmpDir}, mockClient, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check that "Up to date" is shown
	if !strings.Contains(output, "Up to date") {
		t.Errorf("output should indicate up to date, got: %s", output)
	}
}

func TestStatusCommand_MissingMetadata(t *testing.T) {
	// Setup: Create a temp directory without metadata
	tmpDir := t.TempDir()

	mockClient := &indexer.MockClient{
		Stats: &indexer.CollectionStats{
			DocumentCount: 100,
			CollectionName: "swarm-index",
		},
	}

	var buf bytes.Buffer
	err := status.Run([]string{tmpDir}, mockClient, &buf)

	// Should not error, just show appropriate message
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check that path is mentioned with "not indexed" or similar
	if !strings.Contains(output, "Not indexed") && !strings.Contains(output, "not indexed") && !strings.Contains(output, "No metadata") {
		t.Errorf("output should indicate path not indexed, got: %s", output)
	}
}

func TestStatusCommand_TypesenseStats(t *testing.T) {
	tmpDir := t.TempDir()
	meta := &metadata.Metadata{
		LastIndexed: time.Now().Unix(),
		FileCount:   10,
		ContentHash: "abc",
		ProjectType: "go",
		Languages:   []string{"go"},
	}
	writeMetadata(t, tmpDir, meta)

	mockClient := &indexer.MockClient{
		Stats: &indexer.CollectionStats{
			DocumentCount:  45678,
			CollectionName: "swarm-index",
		},
	}

	var buf bytes.Buffer
	err := status.Run([]string{tmpDir}, mockClient, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check Typesense stats are displayed
	if !strings.Contains(output, "45,678") && !strings.Contains(output, "45678") {
		t.Errorf("output should contain document count, got: %s", output)
	}

	if !strings.Contains(output, "swarm-index") {
		t.Errorf("output should contain collection name, got: %s", output)
	}
}

func TestStatusCommand_TypesenseConnectionError(t *testing.T) {
	tmpDir := t.TempDir()
	meta := &metadata.Metadata{
		LastIndexed: time.Now().Unix(),
		FileCount:   10,
		ContentHash: "abc",
		ProjectType: "go",
		Languages:   []string{"go"},
	}
	writeMetadata(t, tmpDir, meta)

	// Mock client that returns error
	mockClient := &indexer.MockClient{
		StatsError: indexer.ErrConnectionFailed,
	}

	var buf bytes.Buffer
	err := status.Run([]string{tmpDir}, mockClient, &buf)

	// Should not error fatally, just show warning
	if err != nil {
		t.Fatalf("should handle connection error gracefully: %v", err)
	}

	output := buf.String()

	// Check that connection error is indicated
	if !strings.Contains(strings.ToLower(output), "error") && !strings.Contains(strings.ToLower(output), "unavailable") && !strings.Contains(strings.ToLower(output), "failed") {
		t.Errorf("output should indicate connection issue, got: %s", output)
	}
}

func TestStatusCommand_MultiplePaths(t *testing.T) {
	// Setup: Create two temp directories with metadata
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	meta1 := &metadata.Metadata{
		LastIndexed: time.Now().Unix(),
		FileCount:   100,
		ContentHash: "hash1",
		ProjectType: "node",
		Languages:   []string{"ts"},
	}
	writeMetadata(t, tmpDir1, meta1)

	meta2 := &metadata.Metadata{
		LastIndexed: time.Now().Unix(),
		FileCount:   200,
		ContentHash: "hash2",
		ProjectType: "go",
		Languages:   []string{"go"},
	}
	writeMetadata(t, tmpDir2, meta2)

	mockClient := &indexer.MockClient{
		Stats: &indexer.CollectionStats{
			DocumentCount:  500,
			CollectionName: "swarm-index",
		},
	}

	var buf bytes.Buffer
	err := status.Run([]string{tmpDir1, tmpDir2}, mockClient, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check both paths are displayed
	if !strings.Contains(output, tmpDir1) {
		t.Errorf("output should contain first path, got: %s", output)
	}
	if !strings.Contains(output, tmpDir2) {
		t.Errorf("output should contain second path, got: %s", output)
	}

	// Check both project types
	if !strings.Contains(output, "node") {
		t.Errorf("output should contain 'node' project type, got: %s", output)
	}
	if !strings.Contains(output, "go") {
		t.Errorf("output should contain 'go' project type, got: %s", output)
	}
}

// Helper to write metadata file
func writeMetadata(t *testing.T, dir string, meta *metadata.Metadata) {
	t.Helper()
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ".swarm-indexer-metadata.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
