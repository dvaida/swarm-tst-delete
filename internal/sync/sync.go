package sync

import "time"

// Sync performs incremental synchronization of files to Typesense.
// It compares file mtimes against indexed timestamps to determine
// which files need to be added, updated, or skipped.
// It also removes documents for files that no longer exist.
func Sync(files []FileInfo, projects []Project, client Client) SyncResult {
	var result SyncResult

	// Build a set of current file paths for cleanup phase
	currentPaths := make(map[string]bool)
	for _, f := range files {
		currentPaths[f.Path] = true
	}

	// Process each file
	for _, file := range files {
		docID := ComputeDocumentID(file.Path)
		existingDoc, err := client.GetDocument(docID)
		if err != nil {
			result.Failed++
			continue
		}

		if existingDoc == nil {
			// New file - add
			doc := Document{
				ID:        docID,
				Path:      file.Path,
				IndexedAt: time.Now().Unix(),
			}
			if err := client.UpsertDocument(doc); err != nil {
				result.Failed++
			} else {
				result.Added++
			}
		} else if file.LastModified > existingDoc.IndexedAt {
			// Modified file - update
			doc := Document{
				ID:        docID,
				Path:      file.Path,
				IndexedAt: time.Now().Unix(),
			}
			if err := client.UpsertDocument(doc); err != nil {
				result.Failed++
			} else {
				result.Updated++
			}
		} else {
			// Unchanged file - skip
			result.Unchanged++
		}
	}

	// Cleanup: find and delete documents for files that no longer exist
	for _, project := range projects {
		docs, err := client.SearchByPathPrefix(project.RootPath)
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
