package chunker

import (
	"regexp"
	"strings"
)

// Language-specific patterns for function/class detection
var (
	goFuncPattern         = regexp.MustCompile(`(?m)^func\s+`)
	pythonDefClassPattern = regexp.MustCompile(`(?m)^(class|def)\s+\w+`)
	jsFuncPattern         = regexp.MustCompile(`(?m)^(async\s+)?function\s+\w+|^(export\s+)?(async\s+)?function\s+\w+|^class\s+\w+`)
	javaMethodPattern     = regexp.MustCompile(`(?m)^\s*(public|private|protected)?\s*(static)?\s*\w+\s+\w+\s*\(`)
)

// ChunkCode splits code content into semantic chunks based on language
func ChunkCode(content string, language string) ([]Chunk, error) {
	if strings.TrimSpace(content) == "" {
		return []Chunk{}, nil
	}

	var pattern *regexp.Regexp
	switch language {
	case "go":
		pattern = goFuncPattern
	case "python":
		pattern = pythonDefClassPattern
	case "javascript", "typescript":
		pattern = jsFuncPattern
	case "java":
		pattern = javaMethodPattern
	default:
		// For unknown languages, return as single chunk
		return []Chunk{{
			Content:   content,
			StartLine: 1,
			EndLine:   strings.Count(content, "\n") + 1,
			ChunkType: "code",
		}}, nil
	}

	return chunkByPattern(content, pattern, language)
}

// chunkByPattern splits content at pattern matches
func chunkByPattern(content string, pattern *regexp.Regexp, language string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	matches := pattern.FindAllStringIndex(content, -1)

	if len(matches) == 0 {
		// No patterns found, return whole content as single chunk
		return []Chunk{{
			Content:   content,
			StartLine: 1,
			EndLine:   len(lines),
			ChunkType: "code",
		}}, nil
	}

	// Convert byte offsets to line numbers
	matchLines := make([]int, len(matches))
	for i, match := range matches {
		matchLines[i] = strings.Count(content[:match[0]], "\n") + 1
	}

	var chunks []Chunk

	// Handle content before first match (imports, package declaration, etc.)
	if matchLines[0] > 1 {
		beforeContent := strings.Join(lines[:matchLines[0]-1], "\n")
		if strings.TrimSpace(beforeContent) != "" {
			chunks = append(chunks, Chunk{
				Content:   beforeContent,
				StartLine: 1,
				EndLine:   matchLines[0] - 1,
				ChunkType: "preamble",
			})
		}
	}

	// Process each match
	for i := 0; i < len(matchLines); i++ {
		startLine := matchLines[i]
		var endLine int
		if i+1 < len(matchLines) {
			endLine = matchLines[i+1] - 1
		} else {
			endLine = len(lines)
		}

		// Trim trailing empty lines
		for endLine > startLine && strings.TrimSpace(lines[endLine-1]) == "" {
			endLine--
		}

		chunkContent := strings.Join(lines[startLine-1:endLine], "\n")
		chunkType := determineChunkType(chunkContent, language)

		chunk := Chunk{
			Content:   chunkContent,
			StartLine: startLine,
			EndLine:   endLine,
			ChunkType: chunkType,
		}

		// Split large chunks
		chunks = append(chunks, splitLargeChunk(chunk)...)
	}

	return chunks, nil
}

// determineChunkType determines the type of chunk based on content and language
func determineChunkType(content string, language string) string {
	switch language {
	case "python":
		if strings.HasPrefix(strings.TrimSpace(content), "class ") {
			return "class"
		}
		return "function"
	case "javascript", "typescript":
		trimmed := strings.TrimSpace(content)
		if strings.HasPrefix(trimmed, "class ") {
			return "class"
		}
		return "function"
	default:
		return "function"
	}
}
