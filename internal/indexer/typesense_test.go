package indexer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewTypesenseClient_Success(t *testing.T) {
	client, err := NewTypesenseClient("http://localhost:8108", "test-api-key", "test-collection")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
}

func TestNewTypesenseClient_EmptyAPIKey(t *testing.T) {
	_, err := NewTypesenseClient("http://localhost:8108", "", "test-collection")
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
	if !strings.Contains(err.Error(), "API key") {
		t.Errorf("error should mention API key, got: %v", err)
	}
}

func TestNewTypesenseClient_EmptyURL(t *testing.T) {
	_, err := NewTypesenseClient("", "test-api-key", "test-collection")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
	if !strings.Contains(err.Error(), "URL") {
		t.Errorf("error should mention URL, got: %v", err)
	}
}

func TestNewTypesenseClient_EmptyCollection(t *testing.T) {
	_, err := NewTypesenseClient("http://localhost:8108", "test-api-key", "")
	if err == nil {
		t.Fatal("expected error for empty collection")
	}
	if !strings.Contains(err.Error(), "collection") {
		t.Errorf("error should mention collection, got: %v", err)
	}
}

func TestEnsureCollection_CreatesIfNotExists(t *testing.T) {
	createdCollection := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/collections/") {
			// Collection doesn't exist
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
			return
		}
		if r.Method == "POST" && r.URL.Path == "/collections" {
			createdCollection = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"name":          "test-collection",
				"num_documents": 0,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewTypesenseClient(server.URL, "test-api-key", "test-collection")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.EnsureCollection(context.Background())
	if err != nil {
		t.Fatalf("EnsureCollection failed: %v", err)
	}

	if !createdCollection {
		t.Error("expected collection to be created")
	}
}

func TestEnsureCollection_AlreadyExists(t *testing.T) {
	collectionRequested := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/collections/") {
			collectionRequested = true
			// Collection exists
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"name":          "test-collection",
				"num_documents": 10,
			})
			return
		}
		if r.Method == "POST" && r.URL.Path == "/collections" {
			t.Error("should not try to create collection when it exists")
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewTypesenseClient(server.URL, "test-api-key", "test-collection")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.EnsureCollection(context.Background())
	if err != nil {
		t.Fatalf("EnsureCollection failed: %v", err)
	}

	if !collectionRequested {
		t.Error("expected collection to be checked")
	}
}

func TestUpsertChunks_SingleChunk(t *testing.T) {
	upsertCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/documents/import") {
			upsertCalled = true
			w.WriteHeader(http.StatusOK)
			// JSONL response for import
			_, _ = w.Write([]byte(`{"success":true}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewTypesenseClient(server.URL, "test-api-key", "test-collection")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	chunks := []IndexedChunk{
		{
			ID:          "test-id-1",
			FilePath:    "/path/to/file.go",
			ProjectPath: "/path/to/project",
			ProjectType: "go",
			Language:    "go",
			ChunkType:   "function",
			Content:     "func main() {}",
			Embedding:   []float32{0.1, 0.2, 0.3},
			StartLine:   1,
			EndLine:     3,
			LastIndexed: 1234567890,
		},
	}

	err = client.UpsertChunks(context.Background(), chunks)
	if err != nil {
		t.Fatalf("UpsertChunks failed: %v", err)
	}

	if !upsertCalled {
		t.Error("expected upsert to be called")
	}
}

func TestUpsertChunks_EmptySlice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request should be made for empty slice")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewTypesenseClient(server.URL, "test-api-key", "test-collection")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.UpsertChunks(context.Background(), []IndexedChunk{})
	if err != nil {
		t.Fatalf("UpsertChunks failed for empty slice: %v", err)
	}
}

func TestUpsertChunks_MultipleBatches(t *testing.T) {
	batchCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/documents/import") {
			batchCount++
			w.WriteHeader(http.StatusOK)
			// Return success for each document in batch
			_, _ = w.Write([]byte(`{"success":true}` + "\n" + `{"success":true}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewTypesenseClient(server.URL, "test-api-key", "test-collection")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Set batch size to 2 for testing
	client.batchSize = 2

	// Create 5 chunks - should result in 3 batches (2, 2, 1)
	chunks := make([]IndexedChunk, 5)
	for i := 0; i < 5; i++ {
		chunks[i] = IndexedChunk{
			ID:          "test-id",
			FilePath:    "/path/to/file.go",
			ProjectPath: "/path/to/project",
			ProjectType: "go",
			Language:    "go",
			ChunkType:   "function",
			Content:     "content",
			Embedding:   []float32{0.1, 0.2},
			StartLine:   1,
			EndLine:     1,
			LastIndexed: 1234567890,
		}
	}

	err = client.UpsertChunks(context.Background(), chunks)
	if err != nil {
		t.Fatalf("UpsertChunks failed: %v", err)
	}

	if batchCount != 3 {
		t.Errorf("expected 3 batches, got %d", batchCount)
	}
}

func TestSearch_ReturnsResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/multi_search") {
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"results": []interface{}{
					map[string]interface{}{
						"hits": []interface{}{
							map[string]interface{}{
								"document": map[string]interface{}{
									"id":           "test-id-1",
									"file_path":    "/path/to/file.go",
									"project_path": "/path/to/project",
									"project_type": "go",
									"language":     "go",
									"chunk_type":   "function",
									"content":      "func main() {}",
									"embedding":    []float32{0.1, 0.2, 0.3},
									"start_line":   float64(1),
									"end_line":     float64(3),
									"last_indexed": float64(1234567890),
								},
							},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewTypesenseClient(server.URL, "test-api-key", "test-collection")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	results, err := client.Search(context.Background(), "main", []float32{0.1, 0.2, 0.3}, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].ID != "test-id-1" {
		t.Errorf("expected ID 'test-id-1', got '%s'", results[0].ID)
	}
	if results[0].FilePath != "/path/to/file.go" {
		t.Errorf("expected FilePath '/path/to/file.go', got '%s'", results[0].FilePath)
	}
}

func TestSearch_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/multi_search") {
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"results": []interface{}{
					map[string]interface{}{
						"hits": []interface{}{},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewTypesenseClient(server.URL, "test-api-key", "test-collection")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	results, err := client.Search(context.Background(), "nonexistent", []float32{0.1, 0.2}, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_TextOnlySearch(t *testing.T) {
	searchPerformed := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/multi_search") {
			searchPerformed = true
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"results": []interface{}{
					map[string]interface{}{
						"hits": []interface{}{},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewTypesenseClient(server.URL, "test-api-key", "test-collection")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Search with nil embedding - text only
	_, err = client.Search(context.Background(), "query", nil, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if !searchPerformed {
		t.Error("expected search to be performed")
	}
}

func TestDeleteByPath_RemovesDocuments(t *testing.T) {
	deleteRequested := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && strings.Contains(r.URL.Path, "/documents") {
			deleteRequested = true
			// Check filter_by parameter
			filterBy := r.URL.Query().Get("filter_by")
			if !strings.Contains(filterBy, "file_path") {
				t.Errorf("expected filter_by to contain file_path, got: %s", filterBy)
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]int{"num_deleted": 3})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewTypesenseClient(server.URL, "test-api-key", "test-collection")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.DeleteByPath(context.Background(), "/path/to/file.go")
	if err != nil {
		t.Fatalf("DeleteByPath failed: %v", err)
	}

	if !deleteRequested {
		t.Error("expected delete to be requested")
	}
}

func TestDeleteByPath_NoMatchingDocs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && strings.Contains(r.URL.Path, "/documents") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]int{"num_deleted": 0})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewTypesenseClient(server.URL, "test-api-key", "test-collection")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Should not return error when no documents match
	err = client.DeleteByPath(context.Background(), "/nonexistent/path.go")
	if err != nil {
		t.Fatalf("DeleteByPath should not fail when no docs match: %v", err)
	}
}

func TestDeleteByPath_EmptyPath(t *testing.T) {
	client, _ := NewTypesenseClient("http://localhost:8108", "test-api-key", "test-collection")
	err := client.DeleteByPath(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}
