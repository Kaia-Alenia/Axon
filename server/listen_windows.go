//go:build windows

package main

import (
	"golang.org/x/sys/windows"
	"syscall"
)

func init() {
	h := windows.Stdout
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err == nil {
		mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
		windows.SetConsoleMode(h, mode)
	}
}

func setReuseAddrControl(network, address string, c syscall.RawConn) error {
	return nil
}
