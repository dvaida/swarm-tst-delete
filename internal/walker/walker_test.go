package walker_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/swarm-indexer/swarm-indexer/internal/walker"
)

// Helper to create a test directory structure
func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create files
	createFile(t, filepath.Join(dir, "file1.txt"), "hello")
	createFile(t, filepath.Join(dir, "file2.go"), "package main")

	// Create subdirectory with files
	subdir := filepath.Join(dir, "subdir")
	os.MkdirAll(subdir, 0755)
	createFile(t, filepath.Join(subdir, "file3.txt"), "nested")

	return dir
}

func createFile(t *testing.T, path, content string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create file %s: %v", path, err)
	}
}

func createDir(t *testing.T, path string) {
	t.Helper()
	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("failed to create dir %s: %v", path, err)
	}
}

func collectFiles(t *testing.T, ch <-chan walker.FileInfo) []walker.FileInfo {
	t.Helper()
	var files []walker.FileInfo
	for f := range ch {
		files = append(files, f)
	}
	return files
}

func TestWalk_BasicDirectory(t *testing.T) {
	dir := setupTestDir(t)

	ch, err := walker.Walk(dir)
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	var files []walker.FileInfo
	for f := range ch {
		files = append(files, f)
	}

	// Should find all 3 files (file1.txt, file2.go, subdir/file3.txt)
	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}

	// Verify files have correct paths
	paths := make(map[string]bool)
	for _, f := range files {
		paths[filepath.Base(f.Path)] = true
	}

	expected := []string{"file1.txt", "file2.go", "file3.txt"}
	for _, name := range expected {
		if !paths[name] {
			t.Errorf("expected to find %s", name)
		}
	}
}

func TestWalk_RespectsGitignore(t *testing.T) {
	dir := t.TempDir()

	// Create .gitignore
	createFile(t, filepath.Join(dir, ".gitignore"), "*.log\nignored/")

	// Create files that should be walked
	createFile(t, filepath.Join(dir, "keep.txt"), "keep me")

	// Create files that should be ignored
	createFile(t, filepath.Join(dir, "debug.log"), "log content")
	createDir(t, filepath.Join(dir, "ignored"))
	createFile(t, filepath.Join(dir, "ignored", "secret.txt"), "secret")

	ch, err := walker.Walk(dir)
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	var files []walker.FileInfo
	for f := range ch {
		files = append(files, f)
	}

	// Should only find keep.txt
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
		for _, f := range files {
			t.Logf("  found: %s", f.Path)
		}
	}

	if len(files) == 1 && filepath.Base(files[0].Path) != "keep.txt" {
		t.Errorf("expected keep.txt, got %s", files[0].Path)
	}
}

func TestWalk_NestedGitignore(t *testing.T) {
	dir := t.TempDir()

	// Root level .gitignore
	createFile(t, filepath.Join(dir, ".gitignore"), "*.log")

	// Create nested directory with its own .gitignore
	nested := filepath.Join(dir, "nested")
	createDir(t, nested)
	createFile(t, filepath.Join(nested, ".gitignore"), "*.tmp")

	// Create files
	createFile(t, filepath.Join(dir, "root.txt"), "root")
	createFile(t, filepath.Join(dir, "root.log"), "ignored by root gitignore")
	createFile(t, filepath.Join(nested, "nested.txt"), "nested")
	createFile(t, filepath.Join(nested, "nested.tmp"), "ignored by nested gitignore")
	createFile(t, filepath.Join(nested, "nested.log"), "also ignored by root gitignore")

	ch, err := walker.Walk(dir)
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	var files []walker.FileInfo
	for f := range ch {
		files = append(files, f)
	}

	// Should find root.txt and nested.txt only
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
		for _, f := range files {
			t.Logf("  found: %s", f.Path)
		}
	}

	paths := make(map[string]bool)
	for _, f := range files {
		paths[filepath.Base(f.Path)] = true
	}

	if !paths["root.txt"] {
		t.Error("expected root.txt")
	}
	if !paths["nested.txt"] {
		t.Error("expected nested.txt")
	}
}

func TestWalk_SkipsHiddenDirectories(t *testing.T) {
	dir := t.TempDir()

	// Create visible file
	createFile(t, filepath.Join(dir, "visible.txt"), "visible")

	// Create hidden directory with file
	hidden := filepath.Join(dir, ".hidden")
	createDir(t, hidden)
	createFile(t, filepath.Join(hidden, "secret.txt"), "hidden content")

	ch, err := walker.Walk(dir)
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	var files []walker.FileInfo
	for f := range ch {
		files = append(files, f)
	}

	// Should only find visible.txt
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
		for _, f := range files {
			t.Logf("  found: %s", f.Path)
		}
	}

	if len(files) == 1 && filepath.Base(files[0].Path) != "visible.txt" {
		t.Errorf("expected visible.txt, got %s", files[0].Path)
	}
}

func TestWalk_AllowsGitDirectory(t *testing.T) {
	dir := t.TempDir()

	// Create .git directory with a config file
	gitDir := filepath.Join(dir, ".git")
	createDir(t, gitDir)
	createFile(t, filepath.Join(gitDir, "config"), "[core]")

	// Create regular file
	createFile(t, filepath.Join(dir, "main.go"), "package main")

	ch, err := walker.Walk(dir)
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	var files []walker.FileInfo
	for f := range ch {
		files = append(files, f)
	}

	// Should find both main.go and .git/config
	paths := make(map[string]bool)
	for _, f := range files {
		// Use relative path from dir
		relPath, _ := filepath.Rel(dir, f.Path)
		paths[relPath] = true
	}

	if !paths["main.go"] {
		t.Error("expected main.go")
	}
	if !paths[filepath.Join(".git", "config")] {
		t.Error("expected .git/config to be accessible")
	}
}

func TestWalk_HandlesSymlinks(t *testing.T) {
	dir := t.TempDir()

	// Create a subdirectory
	subdir := filepath.Join(dir, "subdir")
	createDir(t, subdir)
	createFile(t, filepath.Join(subdir, "file.txt"), "content")

	// Create a symlink that would cause infinite loop if followed
	loopLink := filepath.Join(subdir, "loop")
	err := os.Symlink(dir, loopLink)
	if err != nil {
		t.Skip("symlinks not supported on this platform")
	}

	ch, err := walker.Walk(dir)
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should complete without hanging
	done := make(chan bool)
	go func() {
		var files []walker.FileInfo
		for f := range ch {
			files = append(files, f)
		}
		done <- true
	}()

	select {
	case <-done:
		// Success - walk completed
	case <-time.After(5 * time.Second):
		t.Fatal("Walk appears to be in infinite loop")
	}
}

func TestWalk_HandlesPermissionErrors(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("test cannot run as root")
	}

	dir := t.TempDir()

	// Create accessible file
	createFile(t, filepath.Join(dir, "accessible.txt"), "content")

	// Create restricted directory
	restricted := filepath.Join(dir, "restricted")
	createDir(t, restricted)
	createFile(t, filepath.Join(restricted, "secret.txt"), "secret")
	os.Chmod(restricted, 0000)
	defer os.Chmod(restricted, 0755) // Cleanup

	ch, err := walker.Walk(dir)
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	var files []walker.FileInfo
	for f := range ch {
		files = append(files, f)
	}

	// Should at least find the accessible file and not crash
	found := false
	for _, f := range files {
		if filepath.Base(f.Path) == "accessible.txt" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to find accessible.txt")
	}
}

func TestWalk_NonExistentRoot(t *testing.T) {
	_, err := walker.Walk("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestIsBinary_TextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "text.txt")
	createFile(t, path, "This is plain text content\nwith multiple lines\n")

	isBin, err := walker.IsBinary(path)
	if err != nil {
		t.Fatalf("IsBinary failed: %v", err)
	}

	if isBin {
		t.Error("expected text file to not be binary")
	}
}

func TestIsBinary_BinaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.bin")

	// Create file with null bytes (binary indicator)
	content := []byte("some text\x00with null bytes")
	err := os.WriteFile(path, content, 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	isBin, err := walker.IsBinary(path)
	if err != nil {
		t.Fatalf("IsBinary failed: %v", err)
	}

	if !isBin {
		t.Error("expected file with null bytes to be binary")
	}
}

func TestIsBinary_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	createFile(t, path, "")

	isBin, err := walker.IsBinary(path)
	if err != nil {
		t.Fatalf("IsBinary failed: %v", err)
	}

	if isBin {
		t.Error("expected empty file to not be binary")
	}
}

func TestIsBinary_NonExistent(t *testing.T) {
	_, err := walker.IsBinary("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestFileInfo_Fields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "hello world"
	createFile(t, path, content)

	ch, err := walker.Walk(dir)
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	var info walker.FileInfo
	for f := range ch {
		info = f
		break
	}

	if info.Path != path {
		t.Errorf("expected path %s, got %s", path, info.Path)
	}

	if info.Size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), info.Size)
	}

	if info.IsDir {
		t.Error("expected IsDir to be false for file")
	}

	if info.ModTime.IsZero() {
		t.Error("expected ModTime to be set")
	}
}
