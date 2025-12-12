// Package indexer provides functionality for reading files and indexing them to Typesense.
package indexer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// FileInfo represents a file to be indexed
type FileInfo struct {
	Path         string
	LastModified time.Time
}

// Project represents a detected project with its root path and type
type Project struct {
	Root string
	Type string
}

// Document represents the document structure for Typesense
type Document struct {
	ID           string
	FilePath     string
	FileName     string
	Directory    string
	Content      string
	ProjectRoot  string
	ProjectType  string
	LastModified time.Time
	IndexedAt    time.Time
}

// IndexResult contains the results of an indexing operation
type IndexResult struct {
	Indexed int
	Skipped int
	Failed  int
	Errors  []error
}

// TypesenseClient defines the interface for Typesense operations
type TypesenseClient interface {
	UpsertDocument(doc Document) error
}

// DocumentIDFromPath generates a document ID from a file path using SHA256 hash
func DocumentIDFromPath(path string) string {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:])[:16]
}

// findProject finds the matching project for a file path
func findProject(filePath string, projects []Project) (string, string) {
	for _, project := range projects {
		if strings.HasPrefix(filePath, project.Root) {
			return project.Root, project.Type
		}
	}
	return "", ""
}

// IndexFiles reads files and indexes them to Typesense
func IndexFiles(files []FileInfo, projects []Project, client TypesenseClient) IndexResult {
	var result IndexResult

	for _, file := range files {
		// Read file content
		content, err := os.ReadFile(file.Path)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Errorf("failed to read %s: %w", file.Path, err))
			continue
		}

		// Check UTF-8 validity
		if !utf8.Valid(content) {
			result.Skipped++
			continue
		}

		// Build document
		absPath, err := filepath.Abs(file.Path)
		if err != nil {
			absPath = file.Path
		}

		projectRoot, projectType := findProject(absPath, projects)

		doc := Document{
			ID:           DocumentIDFromPath(absPath),
			FilePath:     absPath,
			FileName:     filepath.Base(absPath),
			Directory:    filepath.Dir(absPath),
			Content:      string(content),
			ProjectRoot:  projectRoot,
			ProjectType:  projectType,
			LastModified: file.LastModified,
			IndexedAt:    time.Now(),
		}

		// Upsert to Typesense
		if err := client.UpsertDocument(doc); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Errorf("failed to index %s: %w", file.Path, err))
			continue
		}

		result.Indexed++
	}

	return result
}
