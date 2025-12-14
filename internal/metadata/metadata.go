package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

const MetadataFileName = ".swarm-indexer-metadata.json"

// Metadata represents the indexer metadata for a directory
type Metadata struct {
	LastIndexed  int64             `json:"last_indexed"`
	FileCount    int               `json:"file_count"`
	ContentHash  string            `json:"content_hash"`
	ProjectType  string            `json:"project_type"`
	Languages    []string          `json:"languages"`
	Dependencies map[string]string `json:"dependencies"`
}

// Manager handles metadata file operations
type Manager struct{}

// New creates a new Manager
func New() *Manager {
	return &Manager{}
}

// Load loads metadata from a directory
func (m *Manager) Load(dir string) (*Metadata, error) {
	path := filepath.Join(dir, MetadataFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No metadata yet
		}
		return nil, err
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// Save saves metadata to a directory
func (m *Manager) Save(dir string, meta *Metadata) error {
	path := filepath.Join(dir, MetadataFileName)
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ComputeContentHash computes a hash of directory contents for change detection
func (m *Manager) ComputeContentHash(dir string) (string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip hidden directories
			if len(d.Name()) > 0 && d.Name()[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip metadata file itself
		if d.Name() == MetadataFileName {
			return nil
		}
		relPath, _ := filepath.Rel(dir, path)
		info, err := d.Info()
		if err != nil {
			return nil
		}
		files = append(files, relPath+":"+string(rune(info.Size()))+":"+info.ModTime().String())
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(files)
	hash := sha256.New()
	for _, f := range files {
		hash.Write([]byte(f))
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// HasChanged checks if directory content has changed since last index
func (m *Manager) HasChanged(dir string, meta *Metadata) (bool, string, error) {
	currentHash, err := m.ComputeContentHash(dir)
	if err != nil {
		return true, "", err
	}
	if meta == nil || meta.ContentHash != currentHash {
		return true, currentHash, nil
	}
	return false, currentHash, nil
}
