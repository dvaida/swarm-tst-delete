package sync

import (
	"errors"
	"testing"
	"time"
)

// MockClient implements Client interface for testing
type MockClient struct {
	documents map[string]Document // id -> document
	searchFn  func(collection, filter string) ([]Document, error)
	getFn     func(collection, id string) (*Document, error)
	upsertFn  func(collection string, doc Document) error
	deleteFn  func(collection, id string) error
}

func NewMockClient() *MockClient {
	return &MockClient{
		documents: make(map[string]Document),
	}
}

func (m *MockClient) GetDocument(collection, id string) (*Document, error) {
	if m.getFn != nil {
		return m.getFn(collection, id)
	}
	doc, ok := m.documents[id]
	if !ok {
		return nil, nil // Not found
	}
	return &doc, nil
}

func (m *MockClient) UpsertDocument(collection string, doc Document) error {
	if m.upsertFn != nil {
		return m.upsertFn(collection, doc)
	}
	m.documents[doc.ID] = doc
	return nil
}

func (m *MockClient) DeleteDocument(collection, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(collection, id)
	}
	delete(m.documents, id)
	return nil
}

func (m *MockClient) SearchDocuments(collection, filter string) ([]Document, error) {
	if m.searchFn != nil {
		return m.searchFn(collection, filter)
	}
	// Return all documents by default
	docs := make([]Document, 0, len(m.documents))
	for _, doc := range m.documents {
		docs = append(docs, doc)
	}
	return docs, nil
}

func TestSync_NewFiles(t *testing.T) {
	client := NewMockClient()
	now := time.Now()

	files := []FileInfo{
		{Path: "/project/file1.go", LastModified: now},
		{Path: "/project/file2.go", LastModified: now},
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

	// Verify documents were created in client
	if len(client.documents) != 2 {
		t.Errorf("expected 2 documents in client, got %d", len(client.documents))
	}
}

func TestSync_UpdatedFiles(t *testing.T) {
	client := NewMockClient()
	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	// Pre-populate with old document
	oldDoc := Document{
		ID:        DocumentID("/project/file1.go"),
		Path:      "/project/file1.go",
		IndexedAt: oldTime.Unix(),
	}
	client.documents[oldDoc.ID] = oldDoc

	files := []FileInfo{
		{Path: "/project/file1.go", LastModified: newTime},
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

	// Verify indexed_at was updated
	doc := client.documents[oldDoc.ID]
	if doc.IndexedAt <= oldTime.Unix() {
		t.Errorf("expected IndexedAt to be updated, got %d", doc.IndexedAt)
	}
}

func TestSync_UnchangedFiles(t *testing.T) {
	client := NewMockClient()
	indexedTime := time.Now()
	fileTime := indexedTime.Add(-1 * time.Hour) // File is older than index

	// Pre-populate with document indexed after file modification
	doc := Document{
		ID:        DocumentID("/project/file1.go"),
		Path:      "/project/file1.go",
		IndexedAt: indexedTime.Unix(),
	}
	client.documents[doc.ID] = doc

	files := []FileInfo{
		{Path: "/project/file1.go", LastModified: fileTime},
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
}

func TestSync_DeletedFiles(t *testing.T) {
	client := NewMockClient()
	now := time.Now()

	// Pre-populate with documents for files that no longer exist
	doc1 := Document{
		ID:        DocumentID("/project/deleted1.go"),
		Path:      "/project/deleted1.go",
		IndexedAt: now.Unix(),
	}
	doc2 := Document{
		ID:        DocumentID("/project/deleted2.go"),
		Path:      "/project/deleted2.go",
		IndexedAt: now.Unix(),
	}
	client.documents[doc1.ID] = doc1
	client.documents[doc2.ID] = doc2

	// Empty file list - all existing docs should be deleted
	files := []FileInfo{}
	projects := []Project{{Root: "/project"}}

	result := Sync(files, projects, client)

	if result.Deleted != 2 {
		t.Errorf("expected Deleted=2, got %d", result.Deleted)
	}

	// Verify documents were deleted from client
	if len(client.documents) != 0 {
		t.Errorf("expected 0 documents in client, got %d", len(client.documents))
	}
}

func TestSync_FailedOperations(t *testing.T) {
	client := NewMockClient()
	now := time.Now()

	// Make upsert fail
	client.upsertFn = func(collection string, doc Document) error {
		return errors.New("upsert failed")
	}

	files := []FileInfo{
		{Path: "/project/file1.go", LastModified: now},
	}
	projects := []Project{{Root: "/project"}}

	result := Sync(files, projects, client)

	if result.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", result.Failed)
	}
	if result.Added != 0 {
		t.Errorf("expected Added=0, got %d", result.Added)
	}
}

func TestSync_MixedOperations(t *testing.T) {
	client := NewMockClient()
	now := time.Now()
	oldTime := now.Add(-2 * time.Hour)
	veryOldTime := now.Add(-24 * time.Hour)

	// Existing unchanged file (indexed after modification)
	unchangedDoc := Document{
		ID:        DocumentID("/project/unchanged.go"),
		Path:      "/project/unchanged.go",
		IndexedAt: now.Unix(),
	}
	client.documents[unchangedDoc.ID] = unchangedDoc

	// Existing file that needs update (modified after indexing)
	updateDoc := Document{
		ID:        DocumentID("/project/update.go"),
		Path:      "/project/update.go",
		IndexedAt: veryOldTime.Unix(),
	}
	client.documents[updateDoc.ID] = updateDoc

	// Existing file that will be deleted (not in new file list)
	deleteDoc := Document{
		ID:        DocumentID("/project/delete.go"),
		Path:      "/project/delete.go",
		IndexedAt: now.Unix(),
	}
	client.documents[deleteDoc.ID] = deleteDoc

	files := []FileInfo{
		{Path: "/project/unchanged.go", LastModified: oldTime},  // older than indexed
		{Path: "/project/update.go", LastModified: now},         // newer than indexed
		{Path: "/project/new.go", LastModified: now},            // new file
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

	// Verify final state: 3 docs (unchanged, update, new) - delete is gone
	if len(client.documents) != 3 {
		t.Errorf("expected 3 documents in client, got %d", len(client.documents))
	}
}

func TestSync_EmptyFileList(t *testing.T) {
	client := NewMockClient()
	now := time.Now()

	// Pre-populate with documents
	for i := 0; i < 3; i++ {
		path := "/project/file" + string(rune('a'+i)) + ".go"
		doc := Document{
			ID:        DocumentID(path),
			Path:      path,
			IndexedAt: now.Unix(),
		}
		client.documents[doc.ID] = doc
	}

	files := []FileInfo{}
	projects := []Project{{Root: "/project"}}

	result := Sync(files, projects, client)

	if result.Deleted != 3 {
		t.Errorf("expected Deleted=3, got %d", result.Deleted)
	}
	if len(client.documents) != 0 {
		t.Errorf("expected 0 documents in client, got %d", len(client.documents))
	}
}

func TestSync_NoExistingDocs(t *testing.T) {
	client := NewMockClient()
	now := time.Now()

	files := []FileInfo{
		{Path: "/project/file1.go", LastModified: now},
		{Path: "/project/file2.go", LastModified: now},
		{Path: "/project/file3.go", LastModified: now},
	}
	projects := []Project{{Root: "/project"}}

	result := Sync(files, projects, client)

	if result.Added != 3 {
		t.Errorf("expected Added=3, got %d", result.Added)
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

func TestDocumentID(t *testing.T) {
	// Same path should produce same ID
	id1 := DocumentID("/project/file.go")
	id2 := DocumentID("/project/file.go")
	if id1 != id2 {
		t.Errorf("expected same ID for same path, got %s and %s", id1, id2)
	}

	// Different paths should produce different IDs
	id3 := DocumentID("/project/other.go")
	if id1 == id3 {
		t.Errorf("expected different IDs for different paths")
	}

	// ID should not be empty
	if id1 == "" {
		t.Error("expected non-empty ID")
	}
}

func TestSync_MultipleProjects(t *testing.T) {
	client := NewMockClient()
	now := time.Now()

	// Documents from two different projects
	doc1 := Document{
		ID:        DocumentID("/project1/file.go"),
		Path:      "/project1/file.go",
		IndexedAt: now.Unix(),
	}
	doc2 := Document{
		ID:        DocumentID("/project2/file.go"),
		Path:      "/project2/file.go",
		IndexedAt: now.Unix(),
	}
	client.documents[doc1.ID] = doc1
	client.documents[doc2.ID] = doc2

	// Only include files from project1
	files := []FileInfo{
		{Path: "/project1/file.go", LastModified: now.Add(-1 * time.Hour)},
	}
	projects := []Project{{Root: "/project1"}, {Root: "/project2"}}

	result := Sync(files, projects, client)

	// project1/file.go should be unchanged
	if result.Unchanged != 1 {
		t.Errorf("expected Unchanged=1, got %d", result.Unchanged)
	}
	// project2/file.go should be deleted (not in file list but under a registered project)
	if result.Deleted != 1 {
		t.Errorf("expected Deleted=1, got %d", result.Deleted)
	}
}

func TestSync_DeleteFailed(t *testing.T) {
	client := NewMockClient()
	now := time.Now()

	// Pre-populate with document
	doc := Document{
		ID:        DocumentID("/project/file.go"),
		Path:      "/project/file.go",
		IndexedAt: now.Unix(),
	}
	client.documents[doc.ID] = doc

	// Make delete fail
	client.deleteFn = func(collection, id string) error {
		return errors.New("delete failed")
	}

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
