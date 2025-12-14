package detector

import (
	"path/filepath"
	"strings"
)

// ProjectInfo contains detected project information
type ProjectInfo struct {
	Type     string   // e.g., "go", "node", "python", "rust"
	Language string   // Primary language
	Languages []string // All detected languages
}

// Detector detects project type and file languages
type Detector struct{}

// New creates a new Detector
func New() *Detector {
	return &Detector{}
}

// DetectProject detects the type of project in a directory
func (d *Detector) DetectProject(path string) (*ProjectInfo, error) {
	return &ProjectInfo{
		Type:      "unknown",
		Language:  "unknown",
		Languages: []string{},
	}, nil
}

// DetectLanguage detects the language of a file based on extension
func (d *Detector) DetectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".md":
		return "markdown"
	case ".txt":
		return "text"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	default:
		return "unknown"
	}
}
