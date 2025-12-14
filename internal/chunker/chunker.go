package chunker

import (
	"path/filepath"
	"strings"
)

const maxChunkSize = 4000

// Chunk represents a semantic chunk of content from a file
type Chunk struct {
	Content   string
	StartLine int
	EndLine   int
	ChunkType string // function, class, paragraph, header, config_key
}

// ChunkFile splits a file into semantic chunks based on its language
func ChunkFile(path string, content string, language string) ([]Chunk, error) {
	if strings.TrimSpace(content) == "" {
		return []Chunk{}, nil
	}

	// Determine language from extension if not provided or unknown
	lang := strings.ToLower(language)
	if lang == "" || lang == "unknown" {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go":
			lang = "go"
		case ".py":
			lang = "python"
		case ".js":
			lang = "javascript"
		case ".ts":
			lang = "typescript"
		case ".java":
			lang = "java"
		case ".md", ".markdown":
			lang = "markdown"
		case ".yaml", ".yml":
			lang = "yaml"
		case ".json":
			lang = "json"
		case ".toml":
			lang = "toml"
		default:
			lang = "text"
		}
	}

	switch lang {
	case "go", "python", "javascript", "typescript", "java":
		return ChunkCode(content, lang)
	case "markdown":
		return ChunkText(content, true)
	case "yaml":
		return chunkYAML(content)
	case "json":
		return chunkJSON(content)
	case "toml":
		return chunkTOML(content)
	default:
		return ChunkText(content, false)
	}
}

// splitLargeChunk splits a chunk into sub-chunks if it exceeds maxChunkSize
func splitLargeChunk(chunk Chunk) []Chunk {
	if len(chunk.Content) <= maxChunkSize {
		return []Chunk{chunk}
	}

	var result []Chunk
	lines := strings.Split(chunk.Content, "\n")
	var currentContent strings.Builder
	currentStart := chunk.StartLine
	currentLine := chunk.StartLine

	for i, line := range lines {
		if currentContent.Len()+len(line)+1 > maxChunkSize && currentContent.Len() > 0 {
			result = append(result, Chunk{
				Content:   currentContent.String(),
				StartLine: currentStart,
				EndLine:   currentLine - 1,
				ChunkType: chunk.ChunkType,
			})
			currentContent.Reset()
			currentStart = currentLine
		}
		if i > 0 {
			currentContent.WriteString("\n")
		}
		currentContent.WriteString(line)
		currentLine++
	}

	if currentContent.Len() > 0 {
		result = append(result, Chunk{
			Content:   currentContent.String(),
			StartLine: currentStart,
			EndLine:   chunk.EndLine,
			ChunkType: chunk.ChunkType,
		})
	}

	return result
}
