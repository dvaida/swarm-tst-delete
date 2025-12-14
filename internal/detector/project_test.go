package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProject_GoProject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "go" {
		t.Errorf("expected Type='go', got '%s'", info.Type)
	}
}

func TestDetectProject_NodeProject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "node" {
		t.Errorf("expected Type='node', got '%s'", info.Type)
	}
}

func TestDetectProject_PythonProject_Requirements(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "python" {
		t.Errorf("expected Type='python', got '%s'", info.Type)
	}
}

func TestDetectProject_PythonProject_Pyproject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[tool.poetry]"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "python" {
		t.Errorf("expected Type='python', got '%s'", info.Type)
	}
}

func TestDetectProject_PythonProject_Setup(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "setup.py"), []byte("setup()"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "python" {
		t.Errorf("expected Type='python', got '%s'", info.Type)
	}
}

func TestDetectProject_RustProject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "rust" {
		t.Errorf("expected Type='rust', got '%s'", info.Type)
	}
}

func TestDetectProject_JavaProject_Maven(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pom.xml"), []byte("<project>"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "java" {
		t.Errorf("expected Type='java', got '%s'", info.Type)
	}
}

func TestDetectProject_JavaProject_Gradle(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "build.gradle"), []byte("plugins {}"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "java" {
		t.Errorf("expected Type='java', got '%s'", info.Type)
	}
}

func TestDetectProject_RubyProject(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Gemfile"), []byte("source 'https://rubygems.org'"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "ruby" {
		t.Errorf("expected Type='ruby', got '%s'", info.Type)
	}
}

func TestDetectProject_UnknownProject(t *testing.T) {
	dir := t.TempDir()

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "unknown" {
		t.Errorf("expected Type='unknown', got '%s'", info.Type)
	}
}

func TestDetectProject_GitVCS(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasVCS {
		t.Error("expected HasVCS=true")
	}
	if info.VCSType != "git" {
		t.Errorf("expected VCSType='git', got '%s'", info.VCSType)
	}
}

func TestDetectProject_SvnVCS(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".svn"), 0755); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasVCS {
		t.Error("expected HasVCS=true")
	}
	if info.VCSType != "svn" {
		t.Errorf("expected VCSType='svn', got '%s'", info.VCSType)
	}
}

func TestDetectProject_HgVCS(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".hg"), 0755); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasVCS {
		t.Error("expected HasVCS=true")
	}
	if info.VCSType != "hg" {
		t.Errorf("expected VCSType='hg', got '%s'", info.VCSType)
	}
}

func TestDetectProject_VSCodeConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".vscode"), 0755); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasIDEConfig {
		t.Error("expected HasIDEConfig=true")
	}
}

func TestDetectProject_IdeaConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".idea"), 0755); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.HasIDEConfig {
		t.Error("expected HasIDEConfig=true")
	}
}

func TestDetectProject_GoModDependencies(t *testing.T) {
	dir := t.TempDir()
	goModContent := `module test

go 1.21

require (
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
)
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "go" {
		t.Errorf("expected Type='go', got '%s'", info.Type)
	}
	if len(info.Dependencies) == 0 {
		t.Error("expected dependencies to be parsed")
	}
	if info.Dependencies["github.com/spf13/cobra"] != "v1.8.0" {
		t.Errorf("expected cobra v1.8.0, got '%s'", info.Dependencies["github.com/spf13/cobra"])
	}
	if info.Dependencies["github.com/stretchr/testify"] != "v1.9.0" {
		t.Errorf("expected testify v1.9.0, got '%s'", info.Dependencies["github.com/stretchr/testify"])
	}
}

func TestDetectProject_PackageJsonDependencies(t *testing.T) {
	dir := t.TempDir()
	packageJsonContent := `{
  "name": "test",
  "dependencies": {
    "express": "^4.18.0",
    "lodash": "4.17.21"
  },
  "devDependencies": {
    "jest": "^29.0.0"
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Type != "node" {
		t.Errorf("expected Type='node', got '%s'", info.Type)
	}
	if len(info.Dependencies) == 0 {
		t.Error("expected dependencies to be parsed")
	}
	if info.Dependencies["express"] != "^4.18.0" {
		t.Errorf("expected express ^4.18.0, got '%s'", info.Dependencies["express"])
	}
	if info.Dependencies["lodash"] != "4.17.21" {
		t.Errorf("expected lodash 4.17.21, got '%s'", info.Dependencies["lodash"])
	}
	if info.Dependencies["jest"] != "^29.0.0" {
		t.Errorf("expected jest ^29.0.0, got '%s'", info.Dependencies["jest"])
	}
}

func TestDetectProject_NonExistentDir(t *testing.T) {
	_, err := DetectProject("/non/existent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent directory")
	}
}
