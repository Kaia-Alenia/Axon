//go:build windows

package main

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

var (
	kernel32Dll                 = windows.NewLazySystemDLL("kernel32.dll")
	user32Dll                   = windows.NewLazySystemDLL("user32.dll")
	procGetConsoleWindow        = kernel32Dll.NewProc("GetConsoleWindow")
	procSetPriorityClass        = kernel32Dll.NewProc("SetPriorityClass")
	procSetThreadExecutionState = kernel32Dll.NewProc("SetThreadExecutionState")
	procSetProcessInformation   = kernel32Dll.NewProc("SetProcessInformation")
	procGetCurrentProcess       = kernel32Dll.NewProc("GetCurrentProcess")
	procIsIconic                = user32Dll.NewProc("IsIconic")
	procShowWindow              = user32Dll.NewProc("ShowWindow")
)

type processPowerThrottlingState struct {
	Version     uint32
	ControlMask uint32
	StateMask   uint32
}

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
		inMode &^= (windows.ENABLE_QUICK_EDIT_MODE | 0x0008 | 0x0010)
		inMode |= windows.ENABLE_EXTENDED_FLAGS
		windows.SetConsoleMode(in, inMode)
	}
	setupWindowsBackgroundExecution()
}

func setupWindowsBackgroundExecution() {
	hProcess, _, _ := procGetCurrentProcess.Call()
	if hProcess != 0 {
		procSetPriorityClass.Call(hProcess, 0x00008000)
		if err := procSetProcessInformation.Find(); err == nil {
			var state processPowerThrottlingState
			state.Version = 1
			state.ControlMask = 0x1 | 0x4
			state.StateMask = 0
			procSetProcessInformation.Call(
				hProcess,
				4,
				uintptr(unsafe.Pointer(&state)),
				uintptr(unsafe.Sizeof(state)),
			)
		}
	}
	procSetThreadExecutionState.Call(0x80000000 | 0x00000001 | 0x00000040)
}

func isConsoleMinimized() bool {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return false
	}
	res, _, _ := procIsIconic.Call(hwnd)
	return res != 0
}

func minimizeConsoleWindow() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		procShowWindow.Call(hwnd, 6)
	}
}

func setReuseAddrControl(network, address string, c syscall.RawConn) error {
	return nil
}
