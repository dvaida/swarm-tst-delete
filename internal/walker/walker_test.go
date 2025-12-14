package walker_test

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/swarm-indexer/swarm-indexer/internal/walker"
)

// collectFiles drains the channel and returns a sorted slice of paths
func collectFiles(ch <-chan walker.FileInfo) []string {
	var paths []string
	for fi := range ch {
		paths = append(paths, fi.Path)
	}
	sort.Strings(paths)
	return paths
}

func TestWalk_BasicDirectory(t *testing.T) {
	// Create temp directory with nested structure
	tmp := t.TempDir()

	// Create files
	files := []string{
		"file1.txt",
		"subdir/file2.txt",
		"subdir/deep/file3.txt",
	}
	for _, f := range files {
		path := filepath.Join(tmp, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	ch, err := walker.Walk(tmp)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	paths := collectFiles(ch)

	// Should have all 3 files
	if len(paths) != 3 {
		t.Errorf("expected 3 files, got %d: %v", len(paths), paths)
	}

	// Verify all expected files are present
	for _, f := range files {
		expected := filepath.Join(tmp, f)
		found := false
		for _, p := range paths {
			if p == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected file %s not found in results", expected)
		}
	}
}

func TestWalk_FileInfoFields(t *testing.T) {
	tmp := t.TempDir()

	content := []byte("hello world")
	testFile := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := walker.Walk(tmp)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	var fi walker.FileInfo
	for f := range ch {
		fi = f
	}

	if fi.Path != testFile {
		t.Errorf("Path = %v, want %v", fi.Path, testFile)
	}
	if fi.Size != int64(len(content)) {
		t.Errorf("Size = %v, want %v", fi.Size, len(content))
	}
	if fi.IsDir {
		t.Error("IsDir should be false for a file")
	}
	if fi.ModTime.IsZero() {
		t.Error("ModTime should not be zero")
	}
}

func TestWalk_RespectsGitignore(t *testing.T) {
	tmp := t.TempDir()

	// Create .gitignore
	gitignore := filepath.Join(tmp, ".gitignore")
	if err := os.WriteFile(gitignore, []byte("*.log\nignored/\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create files - some should be ignored
	testFiles := map[string]bool{
		"keep.txt":           false, // should NOT be ignored
		"debug.log":          true,  // should be ignored
		"ignored/file.txt":   true,  // should be ignored (directory)
		"subdir/another.txt": false, // should NOT be ignored
	}

	for f := range testFiles {
		path := filepath.Join(tmp, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	ch, err := walker.Walk(tmp)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	paths := collectFiles(ch)

	// Check that ignored files are not present
	for f, shouldIgnore := range testFiles {
		fullPath := filepath.Join(tmp, f)
		found := false
		for _, p := range paths {
			if p == fullPath {
				found = true
				break
			}
		}
		if shouldIgnore && found {
			t.Errorf("file %s should be ignored but was found", f)
		}
		if !shouldIgnore && !found {
			t.Errorf("file %s should not be ignored but was not found", f)
		}
	}
}

func TestWalk_NestedGitignore(t *testing.T) {
	tmp := t.TempDir()

	// Root .gitignore ignores *.tmp
	if err := os.WriteFile(filepath.Join(tmp, ".gitignore"), []byte("*.tmp\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create subdir with its own .gitignore that ignores *.bak
	subdir := filepath.Join(tmp, "subdir")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, ".gitignore"), []byte("*.bak\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test files
	testFiles := map[string]bool{
		"root.txt":        false, // keep
		"root.tmp":        true,  // ignored by root .gitignore
		"subdir/sub.txt":  false, // keep
		"subdir/sub.tmp":  true,  // ignored by root .gitignore (inherited)
		"subdir/sub.bak":  true,  // ignored by subdir .gitignore
	}

	for f := range testFiles {
		path := filepath.Join(tmp, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	ch, err := walker.Walk(tmp)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	paths := collectFiles(ch)

	for f, shouldIgnore := range testFiles {
		fullPath := filepath.Join(tmp, f)
		found := false
		for _, p := range paths {
			if p == fullPath {
				found = true
				break
			}
		}
		if shouldIgnore && found {
			t.Errorf("file %s should be ignored but was found", f)
		}
		if !shouldIgnore && !found {
			t.Errorf("file %s should not be ignored but was not found", f)
		}
	}
}

func TestWalk_SkipsHiddenDirectories(t *testing.T) {
	tmp := t.TempDir()

	// Create regular and hidden directories
	dirs := []string{
		"visible",
		".hidden",
		".git", // special case - should be skipped but detectable
	}

	for _, d := range dirs {
		dir := filepath.Join(tmp, d)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		// Create a file in each
		if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	ch, err := walker.Walk(tmp)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	paths := collectFiles(ch)

	// Should have only visible/file.txt
	if len(paths) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(paths), paths)
	}

	expected := filepath.Join(tmp, "visible", "file.txt")
	if len(paths) > 0 && paths[0] != expected {
		t.Errorf("expected %s, got %s", expected, paths[0])
	}
}

func TestWalk_HandlesSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping symlink test on Windows")
	}

	tmp := t.TempDir()

	// Create a directory with a file
	subdir := filepath.Join(tmp, "real")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink to a file
	if err := os.Symlink(filepath.Join(subdir, "file.txt"), filepath.Join(tmp, "link_to_file.txt")); err != nil {
		t.Fatal(err)
	}

	// Create a circular symlink (symlink to parent)
	if err := os.Symlink(tmp, filepath.Join(subdir, "circular")); err != nil {
		t.Fatal(err)
	}

	// Walk should complete without infinite loop
	ch, err := walker.Walk(tmp)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	done := make(chan bool)
	var paths []string
	go func() {
		paths = collectFiles(ch)
		done <- true
	}()

	select {
	case <-done:
		// Good - completed successfully
	case <-time.After(5 * time.Second):
		t.Fatal("Walk() timed out - possible infinite loop from symlinks")
	}

	// Should have at least the real file
	found := false
	for _, p := range paths {
		if p == filepath.Join(subdir, "file.txt") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("real/file.txt not found in results: %v", paths)
	}
}

func TestWalk_PermissionErrors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping permission test on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	tmp := t.TempDir()

	// Create accessible directory with file
	accessible := filepath.Join(tmp, "accessible")
	if err := os.MkdirAll(accessible, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(accessible, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create restricted directory
	restricted := filepath.Join(tmp, "restricted")
	if err := os.MkdirAll(restricted, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(restricted, "secret.txt"), []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(restricted, 0000); err != nil {
		t.Fatal(err)
	}
	// Restore permissions after test
	t.Cleanup(func() {
		os.Chmod(restricted, 0755)
	})

	// Walk should not fail, just skip the restricted directory
	ch, err := walker.Walk(tmp)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	paths := collectFiles(ch)

	// Should have the accessible file
	expected := filepath.Join(accessible, "file.txt")
	found := false
	for _, p := range paths {
		if p == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("accessible/file.txt should be found: %v", paths)
	}
}

func TestWalk_NonExistentPath(t *testing.T) {
	_, err := walker.Walk("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Walk() should return error for non-existent path")
	}
}

func TestIsBinary_TextFile(t *testing.T) {
	tmp := t.TempDir()

	testCases := []struct {
		name    string
		content []byte
	}{
		{"plain text", []byte("Hello, world!")},
		{"go code", []byte("package main\n\nfunc main() {\n\tprintln(\"hi\")\n}")},
		{"utf8 text", []byte("Hello, 世界! 🎉")},
		{"empty lines", []byte("\n\n\n")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(tmp, "test.txt")
			if err := os.WriteFile(path, tc.content, 0644); err != nil {
				t.Fatal(err)
			}

			binary, err := walker.IsBinary(path)
			if err != nil {
				t.Fatalf("IsBinary() error = %v", err)
			}
			if binary {
				t.Errorf("IsBinary() = true for text file %q", tc.name)
			}
		})
	}
}

func TestIsBinary_BinaryFile(t *testing.T) {
	tmp := t.TempDir()

	testCases := []struct {
		name    string
		content []byte
	}{
		{"null at start", append([]byte{0x00}, []byte("text")...)},
		{"null in middle", append([]byte("text"), append([]byte{0x00}, []byte("more")...)...)},
		{"multiple nulls", []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00}}, // PNG-like header
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(tmp, "test.bin")
			if err := os.WriteFile(path, tc.content, 0644); err != nil {
				t.Fatal(err)
			}

			binary, err := walker.IsBinary(path)
			if err != nil {
				t.Fatalf("IsBinary() error = %v", err)
			}
			if !binary {
				t.Errorf("IsBinary() = false for binary file %q", tc.name)
			}
		})
	}
}

func TestIsBinary_NullAfter8KB(t *testing.T) {
	tmp := t.TempDir()

	// Create file with null byte after 8KB - should be detected as text
	content := make([]byte, 9000)
	for i := range content {
		content[i] = 'a'
	}
	content[8500] = 0x00 // null byte after 8KB

	path := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	binary, err := walker.IsBinary(path)
	if err != nil {
		t.Fatalf("IsBinary() error = %v", err)
	}
	if binary {
		t.Error("IsBinary() = true, but null byte is after 8KB check window")
	}
}

func TestIsBinary_EmptyFile(t *testing.T) {
	tmp := t.TempDir()

	path := filepath.Join(tmp, "empty.txt")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	binary, err := walker.IsBinary(path)
	if err != nil {
		t.Fatalf("IsBinary() error = %v", err)
	}
	if binary {
		t.Error("IsBinary() = true for empty file, want false")
	}
}

func TestIsBinary_NonExistentFile(t *testing.T) {
	_, err := walker.IsBinary("/nonexistent/file.txt")
	if err == nil {
		t.Error("IsBinary() should return error for non-existent file")
	}
}
