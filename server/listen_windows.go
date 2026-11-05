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
	in := windows.Stdin
	var inMode uint32
	if err := windows.GetConsoleMode(in, &inMode); err == nil {
		inMode &^= windows.ENABLE_QUICK_EDIT_MODE
		inMode |= windows.ENABLE_EXTENDED_FLAGS
		windows.SetConsoleMode(in, inMode)
	}
}

func setReuseAddrControl(network, address string, c syscall.RawConn) error {
	return nil
}
