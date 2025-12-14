package walker_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dvaida/swarm-indexer/internal/walker"
)

func TestIsBinary_TextFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "text.txt")

	// Create a normal text file
	content := "Hello, this is a text file.\nWith multiple lines.\nNo null bytes here."
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	isBinary, err := walker.IsBinary(filePath)
	if err != nil {
		t.Fatalf("IsBinary() error = %v", err)
	}
	if isBinary {
		t.Error("IsBinary() = true, want false for text file")
	}
}

func TestIsBinary_BinaryFileWithNullBytes(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "binary.bin")

	// Create a binary file with null bytes
	content := []byte("Some text\x00with null\x00bytes inside")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	isBinary, err := walker.IsBinary(filePath)
	if err != nil {
		t.Fatalf("IsBinary() error = %v", err)
	}
	if !isBinary {
		t.Error("IsBinary() = false, want true for binary file")
	}
}

func TestIsBinary_NullByteAfter8KB(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large.bin")

	// Create file with null byte after 8KB - should be detected as text
	// (we only check first 8KB)
	content := make([]byte, 10*1024) // 10KB
	for i := range content {
		content[i] = 'a' // Fill with 'a'
	}
	content[9000] = 0 // Null byte at 9KB (after 8KB threshold)

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	isBinary, err := walker.IsBinary(filePath)
	if err != nil {
		t.Fatalf("IsBinary() error = %v", err)
	}
	// Should be detected as text since null byte is after 8KB
	if isBinary {
		t.Error("IsBinary() = true, want false (null byte is after 8KB)")
	}
}

func TestIsBinary_NullByteWithin8KB(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large_binary.bin")

	// Create file with null byte within first 8KB
	content := make([]byte, 10*1024) // 10KB
	for i := range content {
		content[i] = 'a' // Fill with 'a'
	}
	content[4000] = 0 // Null byte at 4KB (within 8KB threshold)

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	isBinary, err := walker.IsBinary(filePath)
	if err != nil {
		t.Fatalf("IsBinary() error = %v", err)
	}
	if !isBinary {
		t.Error("IsBinary() = false, want true (null byte within 8KB)")
	}
}

func TestIsBinary_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty.txt")

	// Create empty file
	if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	isBinary, err := walker.IsBinary(filePath)
	if err != nil {
		t.Fatalf("IsBinary() error = %v", err)
	}
	// Empty files should be considered text
	if isBinary {
		t.Error("IsBinary() = true, want false for empty file")
	}
}

func TestIsBinary_SmallBinaryFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "small.bin")

	// Create small binary file with null byte (less than 8KB)
	// PNG header followed by null byte (as would appear in real PNG IHDR chunk)
	content := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	isBinary, err := walker.IsBinary(filePath)
	if err != nil {
		t.Fatalf("IsBinary() error = %v", err)
	}
	if !isBinary {
		t.Error("IsBinary() = false, want true for binary file with null bytes")
	}
}

func TestIsBinary_NonExistentFile(t *testing.T) {
	_, err := walker.IsBinary("/nonexistent/file/path")
	if err == nil {
		t.Error("IsBinary() should return error for non-existent file")
	}
}

func TestIsBinary_UTF8WithBOM(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "utf8bom.txt")

	// UTF-8 BOM followed by text
	content := []byte{0xEF, 0xBB, 0xBF}
	content = append(content, []byte("Hello UTF-8")...)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	isBinary, err := walker.IsBinary(filePath)
	if err != nil {
		t.Fatalf("IsBinary() error = %v", err)
	}
	// UTF-8 BOM file should be detected as text
	if isBinary {
		t.Error("IsBinary() = true, want false for UTF-8 BOM file")
	}
}

func TestIsBinary_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := walker.IsBinary(tmpDir)
	// Should return an error for directories
	if err == nil {
		t.Error("IsBinary() should return error for directory")
	}
}
