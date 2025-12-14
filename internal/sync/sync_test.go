package sync

import (
	"errors"
	"testing"
)

// MockClient is a test double for the Client interface
type MockClient struct {
	documents   map[string]Document
	upsertErr   error
	deleteErr   error
	getErr      error
	searchErr   error
	upsertCalls []Document
	deleteCalls []string
}

func NewMockClient() *MockClient {
	return &MockClient{
		documents:   make(map[string]Document),
		upsertCalls: []Document{},
		deleteCalls: []string{},
	}
}

func (m *MockClient) GetDocument(id string) (*Document, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	doc, exists := m.documents[id]
	if !exists {
		return nil, nil
	}
	return &doc, nil
}

func (m *MockClient) UpsertDocument(doc Document) error {
	m.upsertCalls = append(m.upsertCalls, doc)
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.documents[doc.ID] = doc
	return nil
}

func (m *MockClient) DeleteDocument(id string) error {
	m.deleteCalls = append(m.deleteCalls, id)
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.documents, id)
	return nil
}

func (m *MockClient) SearchByPathPrefix(prefix string) ([]Document, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	var results []Document
	for _, doc := range m.documents {
		if len(doc.Path) >= len(prefix) && doc.Path[:len(prefix)] == prefix {
			results = append(results, doc)
		}
	}
	return results, nil
}

func (m *MockClient) AddExistingDocument(doc Document) {
	m.documents[doc.ID] = doc
}

// Test: New files should be added to the index
func Test_Sync_NewFiles(t *testing.T) {
	client := NewMockClient()
	files := []FileInfo{
		{Path: "/project/file1.txt", LastModified: 1000},
		{Path: "/project/file2.txt", LastModified: 2000},
	}
	projects := []Project{
		{ID: "proj1", RootPath: "/project"},
	}

	result := Sync(files, projects, client)

	if result.Added != 2 {
		t.Errorf("Expected Added=2, got %d", result.Added)
	}
	if result.Updated != 0 {
		t.Errorf("Expected Updated=0, got %d", result.Updated)
	}
	if result.Unchanged != 0 {
		t.Errorf("Expected Unchanged=0, got %d", result.Unchanged)
	}
	if len(client.upsertCalls) != 2 {
		t.Errorf("Expected 2 upsert calls, got %d", len(client.upsertCalls))
	}
}

// Test: Modified files (mtime > indexed_at) should be updated
func Test_Sync_ModifiedFiles(t *testing.T) {
	client := NewMockClient()
	// Pre-populate with existing document that has older indexed_at
	existingID := ComputeDocumentID("/project/file1.txt")
	client.AddExistingDocument(Document{
		ID:        existingID,
		Path:      "/project/file1.txt",
		IndexedAt: 500, // indexed at time 500
	})

	files := []FileInfo{
		{Path: "/project/file1.txt", LastModified: 1000}, // modified at 1000 > 500
	}
	projects := []Project{
		{ID: "proj1", RootPath: "/project"},
	}

	result := Sync(files, projects, client)

	if result.Added != 0 {
		t.Errorf("Expected Added=0, got %d", result.Added)
	}
	if result.Updated != 1 {
		t.Errorf("Expected Updated=1, got %d", result.Updated)
	}
	if result.Unchanged != 0 {
		t.Errorf("Expected Unchanged=0, got %d", result.Unchanged)
	}
	if len(client.upsertCalls) != 1 {
		t.Errorf("Expected 1 upsert call, got %d", len(client.upsertCalls))
	}
}

// Test: Unchanged files (mtime <= indexed_at) should be skipped
func Test_Sync_UnchangedFiles(t *testing.T) {
	client := NewMockClient()
	existingID := ComputeDocumentID("/project/file1.txt")
	client.AddExistingDocument(Document{
		ID:        existingID,
		Path:      "/project/file1.txt",
		IndexedAt: 1000,
	})

	files := []FileInfo{
		{Path: "/project/file1.txt", LastModified: 1000}, // same as indexed_at
	}
	projects := []Project{
		{ID: "proj1", RootPath: "/project"},
	}

	result := Sync(files, projects, client)

	if result.Added != 0 {
		t.Errorf("Expected Added=0, got %d", result.Added)
	}
	if result.Updated != 0 {
		t.Errorf("Expected Updated=0, got %d", result.Updated)
	}
	if result.Unchanged != 1 {
		t.Errorf("Expected Unchanged=1, got %d", result.Unchanged)
	}
	if len(client.upsertCalls) != 0 {
		t.Errorf("Expected 0 upsert calls, got %d", len(client.upsertCalls))
	}
}

// Test: Unchanged files with older mtime should also be skipped
func Test_Sync_UnchangedFiles_OlderMtime(t *testing.T) {
	client := NewMockClient()
	existingID := ComputeDocumentID("/project/file1.txt")
	client.AddExistingDocument(Document{
		ID:        existingID,
		Path:      "/project/file1.txt",
		IndexedAt: 1000,
	})

	files := []FileInfo{
		{Path: "/project/file1.txt", LastModified: 500}, // older than indexed_at
	}
	projects := []Project{
		{ID: "proj1", RootPath: "/project"},
	}

	result := Sync(files, projects, client)

	if result.Unchanged != 1 {
		t.Errorf("Expected Unchanged=1, got %d", result.Unchanged)
	}
	if len(client.upsertCalls) != 0 {
		t.Errorf("Expected 0 upsert calls, got %d", len(client.upsertCalls))
	}
}

// Test: Deleted files (in index but not in file list) should be removed
func Test_Sync_DeletedFiles(t *testing.T) {
	client := NewMockClient()
	// Pre-populate with documents for files that no longer exist
	existingID := ComputeDocumentID("/project/deleted.txt")
	client.AddExistingDocument(Document{
		ID:        existingID,
		Path:      "/project/deleted.txt",
		IndexedAt: 500,
	})

	files := []FileInfo{} // No files - the file was deleted
	projects := []Project{
		{ID: "proj1", RootPath: "/project"},
	}

	result := Sync(files, projects, client)

	if result.Deleted != 1 {
		t.Errorf("Expected Deleted=1, got %d", result.Deleted)
	}
	if len(client.deleteCalls) != 1 {
		t.Errorf("Expected 1 delete call, got %d", len(client.deleteCalls))
	}
}

// Test: Mixed scenario with new, modified, unchanged, and deleted files
func Test_Sync_MixedScenario(t *testing.T) {
	client := NewMockClient()

	// Existing doc - unchanged
	unchangedID := ComputeDocumentID("/project/unchanged.txt")
	client.AddExistingDocument(Document{
		ID:        unchangedID,
		Path:      "/project/unchanged.txt",
		IndexedAt: 1000,
	})

	// Existing doc - modified
	modifiedID := ComputeDocumentID("/project/modified.txt")
	client.AddExistingDocument(Document{
		ID:        modifiedID,
		Path:      "/project/modified.txt",
		IndexedAt: 500,
	})

	// Existing doc - will be deleted (not in file list)
	deletedID := ComputeDocumentID("/project/deleted.txt")
	client.AddExistingDocument(Document{
		ID:        deletedID,
		Path:      "/project/deleted.txt",
		IndexedAt: 500,
	})

	files := []FileInfo{
		{Path: "/project/new.txt", LastModified: 2000},       // new file
		{Path: "/project/unchanged.txt", LastModified: 1000}, // unchanged
		{Path: "/project/modified.txt", LastModified: 1000},  // modified (1000 > 500)
	}
	projects := []Project{
		{ID: "proj1", RootPath: "/project"},
	}

	result := Sync(files, projects, client)

	if result.Added != 1 {
		t.Errorf("Expected Added=1, got %d", result.Added)
	}
	if result.Updated != 1 {
		t.Errorf("Expected Updated=1, got %d", result.Updated)
	}
	if result.Unchanged != 1 {
		t.Errorf("Expected Unchanged=1, got %d", result.Unchanged)
	}
	if result.Deleted != 1 {
		t.Errorf("Expected Deleted=1, got %d", result.Deleted)
	}
}

// Test: Empty file list should handle gracefully
func Test_Sync_EmptyFileList(t *testing.T) {
	client := NewMockClient()
	files := []FileInfo{}
	projects := []Project{}

	result := Sync(files, projects, client)

	if result.Added != 0 {
		t.Errorf("Expected Added=0, got %d", result.Added)
	}
	if result.Updated != 0 {
		t.Errorf("Expected Updated=0, got %d", result.Updated)
	}
	if result.Unchanged != 0 {
		t.Errorf("Expected Unchanged=0, got %d", result.Unchanged)
	}
	if result.Deleted != 0 {
		t.Errorf("Expected Deleted=0, got %d", result.Deleted)
	}
	if result.Failed != 0 {
		t.Errorf("Expected Failed=0, got %d", result.Failed)
	}
}

// Test: Client errors should be counted as failures
func Test_Sync_ClientErrors_Upsert(t *testing.T) {
	client := NewMockClient()
	client.upsertErr = errors.New("upsert failed")

	files := []FileInfo{
		{Path: "/project/file1.txt", LastModified: 1000},
	}
	projects := []Project{
		{ID: "proj1", RootPath: "/project"},
	}

	result := Sync(files, projects, client)

	if result.Failed != 1 {
		t.Errorf("Expected Failed=1, got %d", result.Failed)
	}
	if result.Added != 0 {
		t.Errorf("Expected Added=0 on failure, got %d", result.Added)
	}
}

// Test: Client errors on delete should be counted as failures
func Test_Sync_ClientErrors_Delete(t *testing.T) {
	client := NewMockClient()
	client.deleteErr = errors.New("delete failed")

	existingID := ComputeDocumentID("/project/deleted.txt")
	client.AddExistingDocument(Document{
		ID:        existingID,
		Path:      "/project/deleted.txt",
		IndexedAt: 500,
	})

	files := []FileInfo{}
	projects := []Project{
		{ID: "proj1", RootPath: "/project"},
	}

	result := Sync(files, projects, client)

	if result.Failed != 1 {
		t.Errorf("Expected Failed=1, got %d", result.Failed)
	}
	if result.Deleted != 0 {
		t.Errorf("Expected Deleted=0 on failure, got %d", result.Deleted)
	}
}

// Test: Multiple projects should each have their docs checked for deletion
func Test_Sync_MultipleProjects(t *testing.T) {
	client := NewMockClient()

	// Doc in project1
	proj1DocID := ComputeDocumentID("/project1/file.txt")
	client.AddExistingDocument(Document{
		ID:        proj1DocID,
		Path:      "/project1/file.txt",
		IndexedAt: 500,
	})

	// Doc in project2 that will be deleted
	proj2DocID := ComputeDocumentID("/project2/deleted.txt")
	client.AddExistingDocument(Document{
		ID:        proj2DocID,
		Path:      "/project2/deleted.txt",
		IndexedAt: 500,
	})

	files := []FileInfo{
		{Path: "/project1/file.txt", LastModified: 500}, // unchanged
	}
	projects := []Project{
		{ID: "proj1", RootPath: "/project1"},
		{ID: "proj2", RootPath: "/project2"},
	}

	result := Sync(files, projects, client)

	if result.Unchanged != 1 {
		t.Errorf("Expected Unchanged=1, got %d", result.Unchanged)
	}
	if result.Deleted != 1 {
		t.Errorf("Expected Deleted=1, got %d", result.Deleted)
	}
}

// Test: ComputeDocumentID should return consistent IDs
func Test_ComputeDocumentID_Consistent(t *testing.T) {
	id1 := ComputeDocumentID("/project/file.txt")
	id2 := ComputeDocumentID("/project/file.txt")

	if id1 != id2 {
		t.Errorf("Expected same ID for same path, got %s and %s", id1, id2)
	}
}

// Test: ComputeDocumentID should return different IDs for different paths
func Test_ComputeDocumentID_Different(t *testing.T) {
	id1 := ComputeDocumentID("/project/file1.txt")
	id2 := ComputeDocumentID("/project/file2.txt")

	if id1 == id2 {
		t.Errorf("Expected different IDs for different paths, both got %s", id1)
	}
}
