//go:build windows

package main

import (
	"os"
	"os/exec"
)

// restartIntoExe starts the just-installed binary as a new process.
// Windows has no exec-replace, so the caller exits right after this
// returns and the new process takes over. It only returns an error
// when the new process could not be started.
func restartIntoExe(exePath string) error {
	cmd := exec.Command(exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Start()
}
