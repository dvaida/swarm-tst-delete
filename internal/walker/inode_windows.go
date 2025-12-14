//go:build windows

package walker

import "os"

func getInode(info os.FileInfo) uint64 {
	// Windows doesn't have inodes, return 0 to skip inode-based loop detection
	return 0
}
