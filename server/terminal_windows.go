//go:build windows

package main

import (
	"golang.org/x/sys/windows"
)

var procIsZoomed = user32Dll.NewProc("IsZoomed")

// terminalWidth returns the width of the console window in columns,
// for sizing the QR code to the window. It reads the console screen
// buffer size; on failure it falls back to 80.
func terminalWidth() int {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Stdout, &info); err == nil && info.Size.X > 0 {
		return int(info.Size.X)
	}
	return 80
}

// isConsoleMaximized reports whether the console window is maximized
// (IsZoomed on its window handle).
func isConsoleMaximized() bool {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return false
	}
	res, _, _ := procIsZoomed.Call(hwnd)
	return res != 0
}
