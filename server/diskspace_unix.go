//go:build !windows

package main

import (
	"golang.org/x/sys/unix"
)

// freeDiskSpace returns the bytes available to the current user on the
// filesystem containing path. Implemented with unix.Statfs; the
// Windows counterpart lives in diskspace_windows.go.
func freeDiskSpace(path string) (uint64, error) {
	var st unix.Statfs_t
	if err := unix.Statfs(path, &st); err != nil {
		return 0, err
	}
	// Bsize is int64 on Linux and uint32 on macOS; Bavail is uint64
	// on both, and it (not Bfree) is the space a non-root user can use.
	return uint64(st.Bsize) * st.Bavail, nil
}
