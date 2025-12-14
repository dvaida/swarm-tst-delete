package detector

import (
	"path/filepath"
	"strings"
)

var extensionToLanguage = map[string]string{
	".go":   "go",
	".py":   "python",
	".js":   "javascript",
	".jsx":  "javascript",
	".ts":   "typescript",
	".tsx":  "typescript",
	".java": "java",
	".rs":   "rust",
	".rb":   "ruby",
	".c":    "c",
	".h":    "c",
	".cpp":  "cpp",
	".cc":   "cpp",
	".cxx":  "cpp",
	".hpp":  "cpp",
	".md":   "markdown",
	".json": "json",
	".yaml": "yaml",
	".yml":  "yaml",
	".toml": "toml",
}

// DetectLanguage returns the programming language of a file based on its extension.
func DetectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if lang, ok := extensionToLanguage[ext]; ok {
		return lang
	}
	return "unknown"
}
