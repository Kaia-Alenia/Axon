//go:build !windows

package main

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

type winsize struct {
	Row, Col, Xpixel, Ypixel uint16
}

// terminalWidth returns the width of the console in columns, for
// sizing the QR code to the window. It reads the real terminal size
// (ioctl TIOCGWINSZ), falls back to COLUMNS, then to 80.
func terminalWidth() int {
	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno == 0 && ws.Col > 0 {
		return int(ws.Col)
	}
	if c := os.Getenv("COLUMNS"); c != "" {
		if n, err := strconv.Atoi(c); err == nil && n > 0 {
			return n
		}
	}
	return 80
}

// isConsoleMaximized reports whether the console window is maximized.
// Unix terminals have no window-state concept the server can query,
// so this is always false there; the QR size still adapts through
// terminalWidth.
func isConsoleMaximized() bool {
	return false
}
