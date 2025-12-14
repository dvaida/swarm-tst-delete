package sync

import (
	"crypto/sha256"
	"encoding/hex"
)

// ComputeDocumentID generates a consistent document ID from a file path.
func ComputeDocumentID(path string) string {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:])
}
