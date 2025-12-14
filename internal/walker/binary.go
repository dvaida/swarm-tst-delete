package walker

import (
	"io"
	"os"
)

const binaryCheckSize = 8 * 1024 // 8KB

// IsBinary checks if a file is binary by looking for null bytes
// in the first 8KB of the file.
func IsBinary(path string) (bool, error) {
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

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}

	return false, nil
}
