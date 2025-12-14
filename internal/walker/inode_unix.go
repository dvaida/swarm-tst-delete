//go:build !windows

package walker

import (
	"os"
	"syscall"
)

func getInode(info os.FileInfo) uint64 {
	if sys := info.Sys(); sys != nil {
		if stat, ok := sys.(*syscall.Stat_t); ok {
			return stat.Ino
		}
	}
	return 0
}
