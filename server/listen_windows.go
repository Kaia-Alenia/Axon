//go:build windows

package main

import (
	"syscall"
	"golang.org/x/sys/windows"
)

func init() {
	// Enable Virtual Terminal Processing for Windows CMD
	h := windows.Stdout
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err == nil {
		mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
		windows.SetConsoleMode(h, mode)
	}
}

func setReuseAddrControl(network, address string, c syscall.RawConn) error {
	var opErr error
	err := c.Control(func(fd uintptr) {
		opErr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})
	if opErr != nil {
		return opErr
	}
	return err
}
