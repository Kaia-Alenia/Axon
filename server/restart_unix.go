//go:build !windows

package main

import (
	"os"
	"syscall"
)

// restartIntoExe replaces the current process with exePath (no flags),
// so the just-installed binary takes over in place. It only returns
// when the exec fails.
func restartIntoExe(exePath string) error {
	return syscall.Exec(exePath, []string{exePath}, os.Environ())
}
