package sync

// FileInfo represents a file to be potentially indexed
type FileInfo struct {
	Path         string
	LastModified int64 // Unix timestamp
}

// Project represents a registered project/root
type Project struct {
	ID       string
	RootPath string
}

// Document represents a document in Typesense
type Document struct {
	ID        string
	Path      string
	IndexedAt int64 // Unix timestamp when indexed
}

// SyncResult contains statistics about the sync operation
type SyncResult struct {
	Added     int
	Updated   int
	Unchanged int
	Deleted   int
	Failed    int
}

// Client is the interface for Typesense operations
type Client interface {
	GetDocument(id string) (*Document, error)
	UpsertDocument(doc Document) error
	DeleteDocument(id string) error
	SearchByPathPrefix(prefix string) ([]Document, error)
}
