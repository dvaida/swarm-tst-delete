package detector

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ProjectInfo contains information about a detected software project.
type ProjectInfo struct {
	Type         string            // go, node, python, rust, java, ruby, unknown
	HasVCS       bool              // whether the project has version control
	VCSType      string            // git, svn, hg
	HasIDEConfig bool              // whether the project has IDE configuration
	Dependencies map[string]string // name -> version (best effort)
}

// projectMarkers maps marker files to project types
var projectMarkers = map[string]string{
	"go.mod":           "go",
	"package.json":     "node",
	"requirements.txt": "python",
	"pyproject.toml":   "python",
	"setup.py":         "python",
	"Cargo.toml":       "rust",
	"pom.xml":          "java",
	"build.gradle":     "java",
	"Gemfile":          "ruby",
}

// vcsMarkers maps VCS directories to VCS types
var vcsMarkers = map[string]string{
	".git": "git",
	".svn": "svn",
	".hg":  "hg",
}

// ideMarkers lists IDE configuration directories
var ideMarkers = []string{".vscode", ".idea"}

// DetectProject analyzes a directory and returns information about the project.
func DetectProject(dirPath string) (*ProjectInfo, error) {
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, err
	}

	info := &ProjectInfo{
		Type:         "unknown",
		Dependencies: make(map[string]string),
	}

	// Detect project type
	for marker, projectType := range projectMarkers {
		markerPath := filepath.Join(dirPath, marker)
		if _, err := os.Stat(markerPath); err == nil {
			info.Type = projectType
			// Parse dependencies based on project type
			switch marker {
			case "go.mod":
				parseGoModDependencies(markerPath, info.Dependencies)
			case "package.json":
				parsePackageJsonDependencies(markerPath, info.Dependencies)
			}
			break
		}
	}

	// Detect VCS
	for marker, vcsType := range vcsMarkers {
		markerPath := filepath.Join(dirPath, marker)
		if stat, err := os.Stat(markerPath); err == nil && stat.IsDir() {
			info.HasVCS = true
			info.VCSType = vcsType
			break
		}
	}

	// Detect IDE config
	for _, marker := range ideMarkers {
		markerPath := filepath.Join(dirPath, marker)
		if stat, err := os.Stat(markerPath); err == nil && stat.IsDir() {
			info.HasIDEConfig = true
			break
		}
	}

	return info, nil
}

// parseGoModDependencies extracts dependencies from a go.mod file
func parseGoModDependencies(goModPath string, deps map[string]string) {
	file, err := os.Open(goModPath)
	if err != nil {
		return
	}
	defer file.Close()

	inRequire := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for start of require block
		if strings.HasPrefix(line, "require (") {
			inRequire = true
			continue
		}

		// Check for end of require block
		if inRequire && line == ")" {
			inRequire = false
			continue
		}

		// Parse dependencies inside require block
		if inRequire {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := parts[0]
				version := parts[1]
				deps[name] = version
			}
		}

		// Handle single-line require statements
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				name := parts[1]
				version := parts[2]
				deps[name] = version
			}
		}
	}
}

// parsePackageJsonDependencies extracts dependencies from a package.json file
func parsePackageJsonDependencies(packageJsonPath string, deps map[string]string) {
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	for name, version := range pkg.Dependencies {
		deps[name] = version
	}
	for name, version := range pkg.DevDependencies {
		deps[name] = version
	}
}
