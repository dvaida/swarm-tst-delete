package sync

import (
	"strings"
	"time"
)

// FileInfo represents a file to be indexed
type FileInfo struct {
	Path         string
	LastModified time.Time
	Content      string
}

// Project represents a registered project root
type Project struct {
	Root string
}

// Document represents an indexed document in Typesense
type Document struct {
	ID        string
	Path      string
	Content   string
	IndexedAt time.Time
}

// SyncResult contains statistics from a sync operation
type SyncResult struct {
	Added     int
	Updated   int
	Unchanged int
	Deleted   int
	Failed    int
}

// TypesenseClient defines the interface for Typesense operations
type TypesenseClient interface {
	GetDocument(id string) (*Document, error)
	UpsertDocument(doc Document) error
	DeleteDocument(id string) error
	SearchByPathPrefix(prefix string) ([]Document, error)
}

// DocumentIDFromPath generates a document ID from a file path
func DocumentIDFromPath(path string) string {
	// Replace slashes and special chars to create valid Typesense ID
	id := strings.ReplaceAll(path, "/", "_")
	id = strings.ReplaceAll(id, ".", "_")
	return id
}

// Sync synchronizes files with the Typesense index
func Sync(files []FileInfo, projects []Project, client TypesenseClient) SyncResult {
	var result SyncResult

	// Build set of current file paths for deletion check
	currentPaths := make(map[string]bool)
	for _, f := range files {
		currentPaths[f.Path] = true
	}

	// Process each file: add, update, or skip
	for _, file := range files {
		id := DocumentIDFromPath(file.Path)
		existing, err := client.GetDocument(id)
		if err != nil {
			result.Failed++
			continue
		}

		if existing == nil {
			// New file - add it
			doc := Document{
				ID:        id,
				Path:      file.Path,
				Content:   file.Content,
				IndexedAt: time.Now(),
			}
			if err := client.UpsertDocument(doc); err != nil {
				result.Failed++
			} else {
				result.Added++
			}
		} else if file.LastModified.After(existing.IndexedAt) {
			// Modified file - update it
			doc := Document{
				ID:        id,
				Path:      file.Path,
				Content:   file.Content,
				IndexedAt: time.Now(),
			}
			if err := client.UpsertDocument(doc); err != nil {
				result.Failed++
			} else {
				result.Updated++
			}
		} else {
			// Unchanged file - skip it
			result.Unchanged++
		}
	}

	// Process deletions: find indexed docs no longer in file list
	for _, project := range projects {
		docs, err := client.SearchByPathPrefix(project.Root)
		if err != nil {
			continue
		}

		for _, doc := range docs {
			if !currentPaths[doc.Path] {
				if err := client.DeleteDocument(doc.ID); err != nil {
					result.Failed++
				} else {
					result.Deleted++
				}
			}
		}
	}

	return result
}
