package chunker

// Chunk represents a semantic chunk of content
type Chunk struct {
	Content   string
	Type      string // "code", "text", "comment", etc.
	StartLine int
	EndLine   int
}

// Chunker splits content into semantic chunks
type Chunker struct {
	maxChunkSize int
}

// New creates a new Chunker
func New(maxChunkSize int) *Chunker {
	if maxChunkSize <= 0 {
		maxChunkSize = 1000 // default
	}
	return &Chunker{
		maxChunkSize: maxChunkSize,
	}
}

// ChunkContent splits content into semantic chunks based on language
func (c *Chunker) ChunkContent(content string, language string) ([]Chunk, error) {
	if content == "" {
		return nil, nil
	}

	// Simple chunking: return entire content as one chunk
	lines := 1
	for _, ch := range content {
		if ch == '\n' {
			lines++
		}
	}

	return []Chunk{
		{
			Content:   content,
			Type:      "code",
			StartLine: 1,
			EndLine:   lines,
		},
	}, nil
}
