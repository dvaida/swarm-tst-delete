# Plan: Issue #23 - Implement Semantic Chunking for Code and Text

## Overview

Implement a chunker package that splits files into semantic chunks suitable for RAG retrieval. Code files are split by functions/classes, text/docs by sections and paragraphs, and config files by top-level keys.

## Files to Create

```
internal/chunker/chunker.go      # Main orchestration, Chunk struct, ChunkFile function
internal/chunker/code.go         # Code-aware chunking (Go, Python, JS/TS, Java)
internal/chunker/text.go         # Text/docs chunking (markdown, plain text)
internal/chunker/chunker_test.go # Integration tests
```

## Integration Tests to Write

### Test Cases for Code Chunking

1. **TestChunkFile_GoFunctions** - Go file with multiple functions split at `func` boundaries
2. **TestChunkFile_PythonDefClass** - Python file split at `def` and `class` boundaries
3. **TestChunkFile_JavaScript** - JS file split at function declarations and arrow functions
4. **TestChunkFile_TypeScript** - TS file with class methods
5. **TestChunkFile_Java** - Java file with method signatures

### Test Cases for Text Chunking

6. **TestChunkFile_Markdown** - Markdown file split at headers (##, ###)
7. **TestChunkFile_PlainText** - Plain text split at paragraph breaks (blank lines)

### Test Cases for Config Chunking

8. **TestChunkFile_YAML** - YAML file split by top-level keys
9. **TestChunkFile_JSON** - JSON file split by top-level keys
10. **TestChunkFile_TOML** - TOML file split by sections

### Edge Cases

11. **TestChunkFile_LargeFunction** - Functions exceeding 4000 chars split into sub-chunks
12. **TestChunkFile_EmptyFile** - Empty file returns empty slice
13. **TestChunkFile_NoChunkBoundaries** - File without clear boundaries becomes single chunk
14. **TestChunkFile_AccurateLineNumbers** - Verify start_line and end_line are accurate

## Implementation Approach

### Step 1: Define Chunk Struct and Main Entry Point

```go
type Chunk struct {
    Content   string
    StartLine int
    EndLine   int
    ChunkType string // function, class, paragraph, header, config_key
}

func ChunkFile(path string, content string, language string) ([]Chunk, error)
```

- Route to appropriate chunker based on language
- Handle empty content edge case
- Languages: go, python, javascript, typescript, java, markdown, yaml, json, toml, text

### Step 2: Implement Code Chunking

- Use regex patterns to find function/class boundaries
- Go: `func\s+(\w+|(\([^)]+\)\s+\w+))\s*\(`
- Python: `^(def|class)\s+\w+`
- JavaScript/TypeScript: `function\s+\w+`, `const\s+\w+\s*=\s*(\([^)]*\)|[^=])\s*=>`
- Java: `(public|private|protected)?\s*(static)?\s*\w+\s+\w+\s*\(`
- Split at boundaries, preserve context
- Sub-chunk if > 4000 chars

### Step 3: Implement Text Chunking

- Markdown: Split at `^#{1,6}\s` patterns
- Plain text: Split at blank lines (paragraph breaks)
- Sub-chunk large sections

### Step 4: Implement Config Chunking

- YAML: Split at top-level keys (lines starting without indent followed by `:`)
- JSON: Parse and split at top-level keys
- TOML: Split at `[section]` headers

### Step 5: Line Number Tracking

- Track cumulative line count as we split
- Each chunk records its start and end line (1-indexed)

## Acceptance Criteria Mapping

| Criteria | Test |
|----------|------|
| Go files split at function boundaries | TestChunkFile_GoFunctions |
| Python files split at def/class boundaries | TestChunkFile_PythonDefClass |
| Markdown split at headers | TestChunkFile_Markdown |
| Plain text split at paragraph breaks | TestChunkFile_PlainText |
| Chunks have accurate line numbers | TestChunkFile_AccurateLineNumbers |
| Large functions/sections split into sub-chunks | TestChunkFile_LargeFunction |
| Handles files with mixed content gracefully | TestChunkFile_NoChunkBoundaries |

## Dependencies

- No external dependencies needed
- Uses standard library only (strings, regexp)
