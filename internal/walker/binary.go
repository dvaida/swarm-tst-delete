package walker

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const binaryCheckSize = 8 * 1024 // 8KB

// IsBinary checks if a file is binary by looking for null bytes
// in the first 8KB of the file.
func IsBinary(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, fmt.Errorf("path is a directory: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, binaryCheckSize)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}

	// Empty file is not binary
	if n == 0 {
		return false, nil
	}

	// Check for null bytes in the buffer
	return bytes.Contains(buf[:n], []byte{0}), nil
}
