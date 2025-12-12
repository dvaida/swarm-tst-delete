package sync

import (
	"errors"
	"testing"
	"time"
)

// MockTypesenseClient implements TypesenseClient for testing
type MockTypesenseClient struct {
	documents       map[string]*Document
	searchResults   map[string][]Document
	upsertError     error
	deleteError     error
	getError        error
	searchError     error
	upsertCalls     []Document
	deleteCalls     []string
}

func NewMockClient() *MockTypesenseClient {
	return &MockTypesenseClient{
		documents:     make(map[string]*Document),
		searchResults: make(map[string][]Document),
		upsertCalls:   []Document{},
		deleteCalls:   []string{},
	}
}

func (m *MockTypesenseClient) GetDocument(id string) (*Document, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	doc, exists := m.documents[id]
	if !exists {
		return nil, nil
	}
	return doc, nil
}

func (m *MockTypesenseClient) UpsertDocument(doc Document) error {
	if m.upsertError != nil {
		return m.upsertError
	}
	m.upsertCalls = append(m.upsertCalls, doc)
	m.documents[doc.ID] = &doc
	return nil
}

func (m *MockTypesenseClient) DeleteDocument(id string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	m.deleteCalls = append(m.deleteCalls, id)
	delete(m.documents, id)
	return nil
}

func (m *MockTypesenseClient) SearchByPathPrefix(prefix string) ([]Document, error) {
	if m.searchError != nil {
		return nil, m.searchError
	}
	if results, ok := m.searchResults[prefix]; ok {
		return results, nil
	}
	// Return all documents if no specific prefix results configured
	var all []Document
	for _, doc := range m.documents {
		all = append(all, *doc)
	}
	return all, nil
}

func TestSyncNewFiles(t *testing.T) {
	client := NewMockClient()
	now := time.Now()

	files := []FileInfo{
		{Path: "/project/file1.go", LastModified: now, Content: "package main"},
		{Path: "/project/file2.go", LastModified: now, Content: "package util"},
	}
	projects := []Project{{Root: "/project"}}

	result := Sync(files, projects, client)

	if result.Added != 2 {
		t.Errorf("expected Added=2, got %d", result.Added)
	}
	if result.Updated != 0 {
		t.Errorf("expected Updated=0, got %d", result.Updated)
	}
	if result.Unchanged != 0 {
		t.Errorf("expected Unchanged=0, got %d", result.Unchanged)
	}
	if result.Deleted != 0 {
		t.Errorf("expected Deleted=0, got %d", result.Deleted)
	}
	if result.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", result.Failed)
	}
	if len(client.upsertCalls) != 2 {
		t.Errorf("expected 2 upsert calls, got %d", len(client.upsertCalls))
	}
}

func TestSyncUpdatedFiles(t *testing.T) {
	client := NewMockClient()
	oldTime := time.Now().Add(-time.Hour)
	newTime := time.Now()

	// Pre-populate with old documents
	id := DocumentIDFromPath("/project/file1.go")
	client.documents[id] = &Document{
		ID:        id,
		Path:      "/project/file1.go",
		Content:   "old content",
		IndexedAt: oldTime,
	}

	files := []FileInfo{
		{Path: "/project/file1.go", LastModified: newTime, Content: "new content"},
	}
	projects := []Project{{Root: "/project"}}

	result := Sync(files, projects, client)

	if result.Added != 0 {
		t.Errorf("expected Added=0, got %d", result.Added)
	}
	if result.Updated != 1 {
		t.Errorf("expected Updated=1, got %d", result.Updated)
	}
	if result.Unchanged != 0 {
		t.Errorf("expected Unchanged=0, got %d", result.Unchanged)
	}
	if len(client.upsertCalls) != 1 {
		t.Errorf("expected 1 upsert call, got %d", len(client.upsertCalls))
	}
}

func TestSyncUnchangedFiles(t *testing.T) {
	client := NewMockClient()
	fileTime := time.Now().Add(-time.Hour)
	indexedTime := time.Now() // indexed_at is after file mtime

	// Pre-populate with document indexed after file was modified
	id := DocumentIDFromPath("/project/file1.go")
	client.documents[id] = &Document{
		ID:        id,
		Path:      "/project/file1.go",
		Content:   "content",
		IndexedAt: indexedTime,
	}

	files := []FileInfo{
		{Path: "/project/file1.go", LastModified: fileTime, Content: "content"},
	}
	projects := []Project{{Root: "/project"}}

	result := Sync(files, projects, client)

	if result.Added != 0 {
		t.Errorf("expected Added=0, got %d", result.Added)
	}
	if result.Updated != 0 {
		t.Errorf("expected Updated=0, got %d", result.Updated)
	}
	if result.Unchanged != 1 {
		t.Errorf("expected Unchanged=1, got %d", result.Unchanged)
	}
	if len(client.upsertCalls) != 0 {
		t.Errorf("expected 0 upsert calls, got %d", len(client.upsertCalls))
	}
}

func TestSyncDeletedFiles(t *testing.T) {
	client := NewMockClient()
	indexedTime := time.Now()

	// Pre-populate with documents for files that no longer exist
	id1 := DocumentIDFromPath("/project/deleted.go")
	id2 := DocumentIDFromPath("/project/also_deleted.go")
	client.documents[id1] = &Document{
		ID:        id1,
		Path:      "/project/deleted.go",
		Content:   "deleted",
		IndexedAt: indexedTime,
	}
	client.documents[id2] = &Document{
		ID:        id2,
		Path:      "/project/also_deleted.go",
		Content:   "also deleted",
		IndexedAt: indexedTime,
	}

	// Configure search to return these documents for the project prefix
	client.searchResults["/project"] = []Document{
		*client.documents[id1],
		*client.documents[id2],
	}

	// Empty file list - both documents should be deleted
	files := []FileInfo{}
	projects := []Project{{Root: "/project"}}

	result := Sync(files, projects, client)

	if result.Deleted != 2 {
		t.Errorf("expected Deleted=2, got %d", result.Deleted)
	}
	if len(client.deleteCalls) != 2 {
		t.Errorf("expected 2 delete calls, got %d", len(client.deleteCalls))
	}
}

func TestSyncMixedOperations(t *testing.T) {
	client := NewMockClient()
	oldTime := time.Now().Add(-2 * time.Hour)
	indexedTime := time.Now().Add(-time.Hour)
	newTime := time.Now()

	// Setup existing documents
	// File that will be updated (modified after indexing)
	updatedID := DocumentIDFromPath("/project/updated.go")
	client.documents[updatedID] = &Document{
		ID:        updatedID,
		Path:      "/project/updated.go",
		Content:   "old",
		IndexedAt: indexedTime,
	}

	// File that is unchanged (modified before indexing)
	unchangedID := DocumentIDFromPath("/project/unchanged.go")
	client.documents[unchangedID] = &Document{
		ID:        unchangedID,
		Path:      "/project/unchanged.go",
		Content:   "same",
		IndexedAt: indexedTime,
	}

	// File that will be deleted (in index but not in files)
	deletedID := DocumentIDFromPath("/project/deleted.go")
	client.documents[deletedID] = &Document{
		ID:        deletedID,
		Path:      "/project/deleted.go",
		Content:   "gone",
		IndexedAt: indexedTime,
	}

	// Configure search to return all existing docs
	client.searchResults["/project"] = []Document{
		*client.documents[updatedID],
		*client.documents[unchangedID],
		*client.documents[deletedID],
	}

	files := []FileInfo{
		{Path: "/project/new.go", LastModified: newTime, Content: "new file"},                // new
		{Path: "/project/updated.go", LastModified: newTime, Content: "updated content"},     // updated
		{Path: "/project/unchanged.go", LastModified: oldTime, Content: "same"},              // unchanged
	}
	projects := []Project{{Root: "/project"}}

	result := Sync(files, projects, client)

	if result.Added != 1 {
		t.Errorf("expected Added=1, got %d", result.Added)
	}
	if result.Updated != 1 {
		t.Errorf("expected Updated=1, got %d", result.Updated)
	}
	if result.Unchanged != 1 {
		t.Errorf("expected Unchanged=1, got %d", result.Unchanged)
	}
	if result.Deleted != 1 {
		t.Errorf("expected Deleted=1, got %d", result.Deleted)
	}
	if result.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", result.Failed)
	}
}

func TestSyncEmpty(t *testing.T) {
	client := NewMockClient()
	files := []FileInfo{}
	projects := []Project{}

	result := Sync(files, projects, client)

	if result.Added != 0 {
		t.Errorf("expected Added=0, got %d", result.Added)
	}
	if result.Updated != 0 {
		t.Errorf("expected Updated=0, got %d", result.Updated)
	}
	if result.Unchanged != 0 {
		t.Errorf("expected Unchanged=0, got %d", result.Unchanged)
	}
	if result.Deleted != 0 {
		t.Errorf("expected Deleted=0, got %d", result.Deleted)
	}
	if result.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", result.Failed)
	}
}

func TestSyncFailedOperations(t *testing.T) {
	client := NewMockClient()
	client.upsertError = errors.New("upsert failed")
	now := time.Now()

	files := []FileInfo{
		{Path: "/project/file1.go", LastModified: now, Content: "content"},
		{Path: "/project/file2.go", LastModified: now, Content: "content"},
	}
	projects := []Project{{Root: "/project"}}

	result := Sync(files, projects, client)

	if result.Failed != 2 {
		t.Errorf("expected Failed=2, got %d", result.Failed)
	}
	if result.Added != 0 {
		t.Errorf("expected Added=0, got %d", result.Added)
	}
}

func TestSyncDeleteFailures(t *testing.T) {
	client := NewMockClient()
	client.deleteError = errors.New("delete failed")
	indexedTime := time.Now()

	// Pre-populate with document that will fail to delete
	id := DocumentIDFromPath("/project/deleted.go")
	client.documents[id] = &Document{
		ID:        id,
		Path:      "/project/deleted.go",
		Content:   "deleted",
		IndexedAt: indexedTime,
	}

	client.searchResults["/project"] = []Document{*client.documents[id]}

	files := []FileInfo{}
	projects := []Project{{Root: "/project"}}

	result := Sync(files, projects, client)

	if result.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", result.Failed)
	}
	if result.Deleted != 0 {
		t.Errorf("expected Deleted=0, got %d", result.Deleted)
	}
}

func TestDocumentIDFromPath(t *testing.T) {
	// Document IDs should be consistent and valid
	id1 := DocumentIDFromPath("/project/file.go")
	id2 := DocumentIDFromPath("/project/file.go")

	if id1 != id2 {
		t.Errorf("same path should produce same ID: %q vs %q", id1, id2)
	}

	if id1 == "" {
		t.Error("document ID should not be empty")
	}

	// Different paths should produce different IDs
	id3 := DocumentIDFromPath("/project/other.go")
	if id1 == id3 {
		t.Error("different paths should produce different IDs")
	}
}
