//go:build windows

package main

import (
	"golang.org/x/sys/windows"
)

// freeDiskSpace returns the bytes available to the current user on the
// filesystem containing path. Implemented with GetDiskFreeSpaceEx; the
// Unix counterpart lives in diskspace_unix.go.
func freeDiskSpace(path string) (uint64, error) {
	dirPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var freeBytesAvailable uint64
	if err := windows.GetDiskFreeSpaceEx(dirPtr, &freeBytesAvailable, nil, nil); err != nil {
		return 0, err
	}
	return freeBytesAvailable, nil
}
