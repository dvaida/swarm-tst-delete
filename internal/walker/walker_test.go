package walker_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/dvaida/swarm-indexer/internal/walker"
)

// Helper to collect all files from the Walk channel
func collectFiles(ch <-chan walker.FileInfo) []walker.FileInfo {
	var files []walker.FileInfo
	for f := range ch {
		files = append(files, f)
	}
	return files
}

// Helper to get just paths from FileInfo slice
func getPaths(files []walker.FileInfo) []string {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	sort.Strings(paths)
	return paths
}

func TestWalk_BasicDirectoryTraversal(t *testing.T) {
	// Create temp directory with nested structure
	tmpDir := t.TempDir()

	// Create nested structure:
	// tmpDir/
	//   file1.txt
	//   subdir/
	//     file2.txt
	//     nested/
	//       file3.txt
	if err := os.MkdirAll(filepath.Join(tmpDir, "subdir", "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir", "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir", "nested", "file3.txt"), []byte("content3"), 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := walker.Walk(tmpDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	files := collectFiles(ch)
	paths := getPaths(files)

	expected := []string{
		filepath.Join(tmpDir, "file1.txt"),
		filepath.Join(tmpDir, "subdir", "file2.txt"),
		filepath.Join(tmpDir, "subdir", "nested", "file3.txt"),
	}
	sort.Strings(expected)

	if len(paths) != len(expected) {
		t.Errorf("got %d files, want %d", len(paths), len(expected))
		t.Errorf("got: %v", paths)
		t.Errorf("want: %v", expected)
	}

	for i, p := range paths {
		if p != expected[i] {
			t.Errorf("paths[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestWalk_GitignoreAtRootLevel(t *testing.T) {
	tmpDir := t.TempDir()

	// Create structure:
	// tmpDir/
	//   .gitignore (contains "ignored.txt" and "ignored_dir/")
	//   keep.txt
	//   ignored.txt
	//   ignored_dir/
	//     file.txt
	if err := os.MkdirAll(filepath.Join(tmpDir, "ignored_dir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("ignored.txt\nignored_dir/\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "keep.txt"), []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "ignored.txt"), []byte("ignored"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "ignored_dir", "file.txt"), []byte("nested"), 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := walker.Walk(tmpDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	files := collectFiles(ch)
	paths := getPaths(files)

	// Should only find keep.txt and .gitignore itself
	if len(paths) != 2 {
		t.Errorf("got %d files, want 2", len(paths))
		t.Errorf("got: %v", paths)
	}

	for _, p := range paths {
		base := filepath.Base(p)
		if base == "ignored.txt" {
			t.Errorf("should not find ignored.txt")
		}
		if base == "file.txt" {
			t.Errorf("should not find file.txt inside ignored_dir")
		}
	}
}

func TestWalk_NestedGitignoreFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create structure:
	// tmpDir/
	//   .gitignore (contains "*.log")
	//   root.txt
	//   root.log
	//   subdir/
	//     .gitignore (contains "local.txt")
	//     keep.txt
	//     local.txt
	//     sub.log
	if err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("*.log\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "root.txt"), []byte("root"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "root.log"), []byte("log"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir", ".gitignore"), []byte("local.txt\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir", "keep.txt"), []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir", "local.txt"), []byte("local"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir", "sub.log"), []byte("sublog"), 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := walker.Walk(tmpDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	files := collectFiles(ch)
	paths := getPaths(files)

	// Expected: .gitignore (root), root.txt, subdir/.gitignore, subdir/keep.txt
	// Ignored: root.log (root .gitignore), sub.log (root .gitignore), local.txt (subdir .gitignore)
	expectedNames := map[string]bool{
		".gitignore": true,
		"root.txt":   true,
		"keep.txt":   true,
	}
	ignoredNames := map[string]bool{
		"root.log":  true,
		"sub.log":   true,
		"local.txt": true,
	}

	for _, p := range paths {
		base := filepath.Base(p)
		if ignoredNames[base] {
			t.Errorf("should not find ignored file: %s", base)
		}
	}

	// Verify we found expected files
	foundNames := make(map[string]bool)
	for _, p := range paths {
		foundNames[filepath.Base(p)] = true
	}
	for name := range expectedNames {
		if !foundNames[name] {
			t.Errorf("missing expected file: %s", name)
		}
	}
}

func TestWalk_HiddenDirectoriesSkipped(t *testing.T) {
	tmpDir := t.TempDir()

	// Create structure:
	// tmpDir/
	//   visible.txt
	//   .hidden/
	//     secret.txt
	//   .git/
	//     config
	if err := os.MkdirAll(filepath.Join(tmpDir, ".hidden"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte("visible"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden", "secret.txt"), []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".git", "config"), []byte("config"), 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := walker.Walk(tmpDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	files := collectFiles(ch)
	paths := getPaths(files)

	// Should only find visible.txt
	if len(paths) != 1 {
		t.Errorf("got %d files, want 1", len(paths))
		t.Errorf("got: %v", paths)
	}

	if len(paths) > 0 && filepath.Base(paths[0]) != "visible.txt" {
		t.Errorf("expected visible.txt, got %s", paths[0])
	}
}

func TestWalk_SymlinkLoopHandled(t *testing.T) {
	tmpDir := t.TempDir()

	// Create structure:
	// tmpDir/
	//   file.txt
	//   link_to_parent -> tmpDir (creates a loop)
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	// Create symlink loop
	if err := os.Symlink(tmpDir, filepath.Join(tmpDir, "link_to_parent")); err != nil {
		t.Fatal(err)
	}

	ch, err := walker.Walk(tmpDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	// This should complete without hanging
	files := collectFiles(ch)

	// Should find file.txt at least once, shouldn't loop forever
	if len(files) < 1 {
		t.Errorf("expected at least 1 file, got %d", len(files))
	}
}

func TestWalk_PermissionErrorsHandledGracefully(t *testing.T) {
	// Skip on CI where we might be running as root
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	tmpDir := t.TempDir()

	// Create structure:
	// tmpDir/
	//   readable.txt
	//   unreadable_dir/ (no read permission)
	//     hidden.txt
	if err := os.MkdirAll(filepath.Join(tmpDir, "unreadable_dir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "readable.txt"), []byte("readable"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "unreadable_dir", "hidden.txt"), []byte("hidden"), 0644); err != nil {
		t.Fatal(err)
	}
	// Remove read permission from directory
	if err := os.Chmod(filepath.Join(tmpDir, "unreadable_dir"), 0000); err != nil {
		t.Fatal(err)
	}
	// Restore permissions on cleanup
	t.Cleanup(func() {
		_ = os.Chmod(filepath.Join(tmpDir, "unreadable_dir"), 0755)
	})

	ch, err := walker.Walk(tmpDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	// Should complete without error, skipping unreadable dir
	files := collectFiles(ch)
	paths := getPaths(files)

	// Should at least find readable.txt
	found := false
	for _, p := range paths {
		if filepath.Base(p) == "readable.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find readable.txt")
	}
}

func TestWalk_NonExistentDirectory(t *testing.T) {
	ch, err := walker.Walk("/nonexistent/path/that/does/not/exist")
	if err == nil {
		// If no error, the channel should be empty or we should get an error
		files := collectFiles(ch)
		if len(files) > 0 {
			t.Error("expected error or empty result for non-existent path")
		}
	}
	// Having an error is the expected behavior
}

func TestWalk_FileInfoFields(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("test content here")
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := walker.Walk(tmpDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	files := collectFiles(ch)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if f.Path != filePath {
		t.Errorf("Path = %q, want %q", f.Path, filePath)
	}
	if f.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", f.Size, len(content))
	}
	if f.ModTime.IsZero() {
		t.Error("ModTime should not be zero")
	}
	if f.IsDir {
		t.Error("IsDir should be false for a file")
	}
}

func TestWalk_GlobPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Test glob patterns like *.txt
	// Create structure:
	// tmpDir/
	//   .gitignore (contains "*.tmp")
	//   keep.txt
	//   remove.tmp
	//   subdir/
	//     also.tmp
	if err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("*.tmp\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "keep.txt"), []byte("keep"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "remove.tmp"), []byte("remove"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir", "also.tmp"), []byte("also"), 0644); err != nil {
		t.Fatal(err)
	}

	ch, err := walker.Walk(tmpDir)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	files := collectFiles(ch)

	for _, f := range files {
		if filepath.Ext(f.Path) == ".tmp" {
			t.Errorf("should not find .tmp file: %s", f.Path)
		}
	}
}
