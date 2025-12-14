package chunker

import (
	"regexp"
	"strings"
)

var markdownHeaderPattern = regexp.MustCompile(`(?m)^#{1,6}\s+`)

// ChunkText splits text content into semantic chunks
// If isMarkdown is true, splits at headers; otherwise splits at paragraph breaks
func ChunkText(content string, isMarkdown bool) ([]Chunk, error) {
	if strings.TrimSpace(content) == "" {
		return []Chunk{}, nil
	}

	if isMarkdown {
		return chunkMarkdown(content)
	}
	return chunkPlainText(content)
}

// chunkMarkdown splits markdown content at header boundaries
func chunkMarkdown(content string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	matches := markdownHeaderPattern.FindAllStringIndex(content, -1)

	if len(matches) == 0 {
		// No headers found, return as single chunk
		return []Chunk{{
			Content:   content,
			StartLine: 1,
			EndLine:   len(lines),
			ChunkType: "paragraph",
		}}, nil
	}

	// Convert byte offsets to line numbers
	matchLines := make([]int, len(matches))
	for i, match := range matches {
		matchLines[i] = strings.Count(content[:match[0]], "\n") + 1
	}

	var chunks []Chunk

	// Process each header section
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

		chunk := Chunk{
			Content:   chunkContent,
			StartLine: startLine,
			EndLine:   endLine,
			ChunkType: "header",
		}

		chunks = append(chunks, splitLargeChunk(chunk)...)
	}

	return chunks, nil
}

// chunkPlainText splits plain text at paragraph breaks (blank lines)
func chunkPlainText(content string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")

	var chunks []Chunk
	var currentChunk []string
	currentStart := 1

	for i, line := range lines {
		lineNum := i + 1

		if strings.TrimSpace(line) == "" {
			// Found a blank line
			if len(currentChunk) > 0 {
				chunk := Chunk{
					Content:   strings.Join(currentChunk, "\n"),
					StartLine: currentStart,
					EndLine:   lineNum - 1,
					ChunkType: "paragraph",
				}
				chunks = append(chunks, splitLargeChunk(chunk)...)
				currentChunk = nil
			}
			currentStart = lineNum + 1
		} else {
			currentChunk = append(currentChunk, line)
		}
	}

	// Handle remaining content
	if len(currentChunk) > 0 {
		chunk := Chunk{
			Content:   strings.Join(currentChunk, "\n"),
			StartLine: currentStart,
			EndLine:   len(lines),
			ChunkType: "paragraph",
		}
		chunks = append(chunks, splitLargeChunk(chunk)...)
	}

	// If no chunks were created but content exists, return it as one chunk
	if len(chunks) == 0 && strings.TrimSpace(content) != "" {
		return []Chunk{{
			Content:   content,
			StartLine: 1,
			EndLine:   len(lines),
			ChunkType: "paragraph",
		}}, nil
	}

	return chunks, nil
}

// chunkYAML splits YAML content by top-level keys
func chunkYAML(content string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	topLevelKeyPattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*:`)

	var chunks []Chunk
	var currentChunk []string
	var currentStart int

	for i, line := range lines {
		lineNum := i + 1

		if topLevelKeyPattern.MatchString(line) {
			// Found a top-level key
			if len(currentChunk) > 0 {
				chunk := Chunk{
					Content:   strings.Join(currentChunk, "\n"),
					StartLine: currentStart,
					EndLine:   lineNum - 1,
					ChunkType: "config_key",
				}
				chunks = append(chunks, splitLargeChunk(chunk)...)
			}
			currentChunk = []string{line}
			currentStart = lineNum
		} else if currentStart > 0 {
			currentChunk = append(currentChunk, line)
		}
	}

	// Handle remaining content
	if len(currentChunk) > 0 {
		endLine := len(lines)
		// Trim trailing empty lines
		for endLine > currentStart && strings.TrimSpace(lines[endLine-1]) == "" {
			endLine--
		}
		chunk := Chunk{
			Content:   strings.Join(currentChunk, "\n"),
			StartLine: currentStart,
			EndLine:   endLine,
			ChunkType: "config_key",
		}
		chunks = append(chunks, splitLargeChunk(chunk)...)
	}

	return chunks, nil
}

// chunkJSON splits JSON content by top-level keys
func chunkJSON(content string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	topLevelKeyPattern := regexp.MustCompile(`^\s*"[^"]+"\s*:`)

	var chunks []Chunk
	var currentChunk []string
	var currentStart int
	braceDepth := 0
	bracketDepth := 0

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Track depth for nested structures
		for _, c := range line {
			switch c {
			case '{':
				braceDepth++
			case '}':
				braceDepth--
			case '[':
				bracketDepth++
			case ']':
				bracketDepth--
			}
		}

		// Only match top-level keys (depth 1)
		if braceDepth == 1 && bracketDepth == 0 && topLevelKeyPattern.MatchString(trimmed) {
			// Found a top-level key
			if len(currentChunk) > 0 {
				chunkContent := strings.Join(currentChunk, "\n")
				// Remove trailing comma if present
				chunkContent = strings.TrimSuffix(strings.TrimSpace(chunkContent), ",")
				chunk := Chunk{
					Content:   chunkContent,
					StartLine: currentStart,
					EndLine:   lineNum - 1,
					ChunkType: "config_key",
				}
				chunks = append(chunks, splitLargeChunk(chunk)...)
			}
			currentChunk = []string{line}
			currentStart = lineNum
		} else if currentStart > 0 && trimmed != "{" && trimmed != "}" {
			currentChunk = append(currentChunk, line)
		}
	}

	// Handle remaining content
	if len(currentChunk) > 0 {
		endLine := len(lines)
		// Trim trailing empty lines and closing braces
		for endLine > currentStart {
			trimmed := strings.TrimSpace(lines[endLine-1])
			if trimmed == "" || trimmed == "}" || trimmed == "}," {
				endLine--
			} else {
				break
			}
		}
		chunkContent := strings.Join(currentChunk, "\n")
		chunkContent = strings.TrimSuffix(strings.TrimSpace(chunkContent), ",")
		chunk := Chunk{
			Content:   chunkContent,
			StartLine: currentStart,
			EndLine:   endLine,
			ChunkType: "config_key",
		}
		chunks = append(chunks, splitLargeChunk(chunk)...)
	}

	return chunks, nil
}

// chunkTOML splits TOML content by sections
func chunkTOML(content string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	sectionPattern := regexp.MustCompile(`^\s*\[[^\]]+\]`)

	var chunks []Chunk
	var currentChunk []string
	var currentStart int
	foundFirstSection := false

	for i, line := range lines {
		lineNum := i + 1

		if sectionPattern.MatchString(line) {
			// Found a section header
			if len(currentChunk) > 0 {
				endLine := lineNum - 1
				// Trim trailing empty lines
				for endLine > currentStart && strings.TrimSpace(lines[endLine-1]) == "" {
					endLine--
				}
				chunk := Chunk{
					Content:   strings.Join(currentChunk, "\n"),
					StartLine: currentStart,
					EndLine:   endLine,
					ChunkType: "config_key",
				}
				chunks = append(chunks, splitLargeChunk(chunk)...)
			}
			currentChunk = []string{line}
			currentStart = lineNum
			foundFirstSection = true
		} else if foundFirstSection {
			currentChunk = append(currentChunk, line)
		} else if strings.TrimSpace(line) != "" {
			// Content before first section
			if len(currentChunk) == 0 {
				currentStart = lineNum
			}
			currentChunk = append(currentChunk, line)
		}
	}

	// Handle remaining content
	if len(currentChunk) > 0 {
		endLine := len(lines)
		// Trim trailing empty lines
		for endLine > currentStart && strings.TrimSpace(lines[endLine-1]) == "" {
			endLine--
		}
		chunk := Chunk{
			Content:   strings.Join(currentChunk, "\n"),
			StartLine: currentStart,
			EndLine:   endLine,
			ChunkType: "config_key",
		}
		chunks = append(chunks, splitLargeChunk(chunk)...)
	}

	return chunks, nil
}
