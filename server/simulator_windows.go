//go:build windows

package main

import (
	"syscall"
	"time"
	"unsafe"
)

type MouseInput struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type KeyboardInput struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type InputUnion struct {
	Type uint32
	Mi   MouseInput
	Pad  [8]byte
}

type InputKeyboardStruct struct {
	Type uint32
	Ki   KeyboardInput
	Pad  [8]byte
}

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	procSendInput = user32.NewProc("SendInput")
)

type WindowsSimulator struct{}

func initSimulator() InputSimulator {
	return &WindowsSimulator{}
}

func (s *WindowsSimulator) sendMouseInput(dwFlags uint32, dx, dy int32, mouseData uint32) {
	var input InputUnion
	input.Type = 0
	input.Mi.DwFlags = dwFlags
	input.Mi.Dx = dx
	input.Mi.Dy = dy
	input.Mi.MouseData = mouseData
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(unsafe.Sizeof(input)))
}

func (s *WindowsSimulator) sendKeyboardInput(dwFlags uint32, wVk, wScan uint16) {
	var input InputKeyboardStruct
	input.Type = 1
	input.Ki.DwFlags = dwFlags
	input.Ki.WVk = wVk
	input.Ki.WScan = wScan
	procSendInput.Call(1, uintptr(unsafe.Pointer(&input)), uintptr(unsafe.Sizeof(input)))
}

func (s *WindowsSimulator) MoveMouse(dx, dy float64) {
	s.sendMouseInput(0x0001, int32(dx), int32(dy), 0)
}

func (s *WindowsSimulator) Click(button string) {
	s.MouseDown(button)
	time.Sleep(10 * time.Millisecond)
	s.MouseUp(button)
}

func (s *WindowsSimulator) MouseDown(button string) {
	if button == "right" {
		s.sendMouseInput(0x0008, 0, 0, 0)
	} else if button == "middle" {
		s.sendMouseInput(0x0020, 0, 0, 0)
	} else {
		s.sendMouseInput(0x0002, 0, 0, 0)
	}
}

func (s *WindowsSimulator) MouseUp(button string) {
	if button == "right" {
		s.sendMouseInput(0x0010, 0, 0, 0)
	} else if button == "middle" {
		s.sendMouseInput(0x0040, 0, 0, 0)
	} else {
		s.sendMouseInput(0x0004, 0, 0, 0)
	}
}

func (s *WindowsSimulator) Scroll(direction string) {
	var data uint32 = 120
	if direction == "down" {
		data = 4294967176
	}
	s.sendMouseInput(0x0800, 0, 0, data)
}

func (s *WindowsSimulator) Type(text string) {
	for _, r := range text {
		s.sendKeyboardInput(0x0004, 0, uint16(r))
		s.sendKeyboardInput(0x0004|0x0002, 0, uint16(r))
	}
}

func (s *WindowsSimulator) Key(key string) {
	var vk uint16
	switch key {
	case "BackSpace":
		vk = 0x08
	case "Return":
		vk = 0x0D
	case "Tab":
		vk = 0x09
	case "Escape":
		vk = 0x1B
	case "space":
		vk = 0x20
	case "Up":
		vk = 0x26
	case "Down":
		vk = 0x28
	case "Left":
		vk = 0x25
	case "Right":
		vk = 0x27
	case "Home":
		vk = 0x24
	case "End":
		vk = 0x23
	case "Prior":
		vk = 0x21
	case "Next":
		vk = 0x22
	case "Delete":
		vk = 0x2E
	case "F1":
		vk = 0x70
	case "F2":
		vk = 0x71
	case "F3":
		vk = 0x72
	case "F4":
		vk = 0x73
	case "F5":
		vk = 0x74
	case "F6":
		vk = 0x75
	case "F7":
		vk = 0x76
	case "F8":
		vk = 0x77
	case "F9":
		vk = 0x78
	case "F10":
		vk = 0x79
	case "F11":
		vk = 0x7A
	case "F12":
		vk = 0x7B
	default:
		return
	}
	s.sendKeyboardInput(0, vk, 0)
	s.sendKeyboardInput(0x0002, vk, 0)
}

func (s *WindowsSimulator) KeyCombo(modifier, key string) {
	var modVk uint16
	switch modifier {
	case "ctrl":
		modVk = 0x11
	case "alt":
		modVk = 0x12
	case "shift":
		modVk = 0x10
	default:
		return
	}
	var keyVk uint16
	if len(key) == 1 {
		char := []rune(key)[0]
		if char >= 'a' && char <= 'z' {
			keyVk = uint16(char - 32)
		} else if char >= 'A' && char <= 'Z' {
			keyVk = uint16(char)
		} else {
			keyVk = uint16(char)
		}
	} else {
		return
	}
	s.sendKeyboardInput(0, modVk, 0)
	s.sendKeyboardInput(0, keyVk, 0)
	s.sendKeyboardInput(0x0002, keyVk, 0)
	s.sendKeyboardInput(0x0002, modVk, 0)
}

func (s *WindowsSimulator) Close() {}

