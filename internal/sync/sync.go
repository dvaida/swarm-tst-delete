package sync

import (
	"encoding/base64"
	"strings"
	"time"
)

// FileInfo represents a file to be synced
type FileInfo struct {
	Path         string
	LastModified time.Time
}

// Project represents a project root for filtering
type Project struct {
	Root string
}

// Document represents a document in Typesense
type Document struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	IndexedAt int64  `json:"indexed_at"`
}

// SyncResult contains statistics from a sync operation
type SyncResult struct {
	Added     int
	Updated   int
	Unchanged int
	Deleted   int
	Failed    int
}

// Client interface for Typesense operations
type Client interface {
	GetDocument(collection, id string) (*Document, error)
	UpsertDocument(collection string, doc Document) error
	DeleteDocument(collection, id string) error
	SearchDocuments(collection, filter string) ([]Document, error)
}

const collection = "files"

// DocumentID computes a unique document ID from a file path
func DocumentID(path string) string {
	return base64.URLEncoding.EncodeToString([]byte(path))
}

// Sync performs incremental synchronization of files to Typesense
func Sync(files []FileInfo, projects []Project, client Client) SyncResult {
	var result SyncResult

	// Build set of current file paths for quick lookup
	currentFiles := make(map[string]bool)
	for _, f := range files {
		currentFiles[f.Path] = true
	}

	// Process each file
	for _, file := range files {
		id := DocumentID(file.Path)
		existing, err := client.GetDocument(collection, id)
		if err != nil {
			result.Failed++
			continue
		}

		now := time.Now().Unix()

		if existing == nil {
			// New file - add
			doc := Document{
				ID:        id,
				Path:      file.Path,
				IndexedAt: now,
			}
			if err := client.UpsertDocument(collection, doc); err != nil {
				result.Failed++
			} else {
				result.Added++
			}
		} else if file.LastModified.Unix() > existing.IndexedAt {
			// File modified after last index - update
			doc := Document{
				ID:        id,
				Path:      file.Path,
				IndexedAt: now,
			}
			if err := client.UpsertDocument(collection, doc); err != nil {
				result.Failed++
			} else {
				result.Updated++
			}
		} else {
			// File unchanged
			result.Unchanged++
		}
	}

	// Find and delete documents for files that no longer exist
	for _, project := range projects {
		docs, err := client.SearchDocuments(collection, "path:"+project.Root+"*")
		if err != nil {
			continue
		}

		for _, doc := range docs {
			// Check if this document's path is under any project root
			underProject := false
			for _, p := range projects {
				if strings.HasPrefix(doc.Path, p.Root) {
					underProject = true
					break
				}
			}

			if underProject && !currentFiles[doc.Path] {
				if err := client.DeleteDocument(collection, doc.ID); err != nil {
					result.Failed++
				} else {
					result.Deleted++
				}
			}
		}
	}

	return result
}
