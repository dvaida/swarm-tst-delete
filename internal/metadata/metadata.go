package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const MetadataFileName = ".swarm-indexer-metadata.json"

// Metadata represents the indexing metadata stored in .swarm-indexer-metadata.json
type Metadata struct {
	LastIndexed  int64             `json:"last_indexed"`
	FileCount    int               `json:"file_count"`
	ContentHash  string            `json:"content_hash"`
	ProjectType  string            `json:"project_type"`
	Languages    []string          `json:"languages"`
	Dependencies map[string]string `json:"dependencies"`
}

// Load reads the metadata file from the given path
func Load(path string) (*Metadata, error) {
	metaPath := filepath.Join(path, MetadataFileName)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// ComputeContentHash computes a hash of all files in the given directory
func ComputeContentHash(path string) (string, error) {
	h := sha256.New()

	// Collect all file paths first for deterministic ordering
	var files []string
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the metadata file itself and directories
		if info.IsDir() || filepath.Base(p) == MetadataFileName {
			return nil
		}
		files = append(files, p)
		return nil
	})
	if err != nil {
		return "", err
	}

	// Sort for deterministic hash
	sort.Strings(files)

	// Hash each file's path and content
	for _, p := range files {
		// Include relative path in hash
		rel, err := filepath.Rel(path, p)
		if err != nil {
			return "", err
		}
		h.Write([]byte(rel))

		// Include file content in hash
		f, err := os.Open(p)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(h, f); err != nil {
			f.Close()
			return "", err
		}
		f.Close()
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
