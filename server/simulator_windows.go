//go:build windows

package main

import (
	"encoding/binary"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type rawInput [40]byte

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	procSendInput = user32.NewProc("SendInput")
)

type WindowsSimulator struct {
	mu   sync.Mutex
	remX float64
	remY float64
}

func initSimulator() InputSimulator {
	return &WindowsSimulator{}
}
func makeMouseInput(dwFlags uint32, dx, dy int32, mouseData uint32) rawInput {
	var in rawInput
	binary.LittleEndian.PutUint32(in[0:4], 0)
	binary.LittleEndian.PutUint32(in[8:12], uint32(dx))
	binary.LittleEndian.PutUint32(in[12:16], uint32(dy))
	binary.LittleEndian.PutUint32(in[16:20], mouseData)
	binary.LittleEndian.PutUint32(in[20:24], dwFlags)
	return in
}
func makeKeyboardInput(dwFlags uint32, wVk, wScan uint16) rawInput {
	var in rawInput
	binary.LittleEndian.PutUint32(in[0:4], 1)
	binary.LittleEndian.PutUint16(in[8:10], wVk)
	binary.LittleEndian.PutUint16(in[10:12], wScan)
	binary.LittleEndian.PutUint32(in[12:16], dwFlags)
	return in
}
func (s *WindowsSimulator) sendMouseInput(dwFlags uint32, dx, dy int32, mouseData uint32) {
	in := makeMouseInput(dwFlags, dx, dy, mouseData)
	procSendInput.Call(1, uintptr(unsafe.Pointer(&in[0])), 40)
}
func (s *WindowsSimulator) sendKeyboardInput(dwFlags uint32, wVk, wScan uint16) {
	in := makeKeyboardInput(dwFlags, wVk, wScan)
	procSendInput.Call(1, uintptr(unsafe.Pointer(&in[0])), 40)
}
func (s *WindowsSimulator) sendInputs(inputs []rawInput) {
	if len(inputs) == 0 {
		return
	}
	procSendInput.Call(uintptr(len(inputs)), uintptr(unsafe.Pointer(&inputs[0])), 40)
}
func (s *WindowsSimulator) MoveMouse(dx, dy float64) {
	s.mu.Lock()
	s.remX += dx
	s.remY += dy
	ix := int32(s.remX)
	iy := int32(s.remY)
	if ix == 0 && iy == 0 {
		s.mu.Unlock()
		return
	}
	s.remX -= float64(ix)
	s.remY -= float64(iy)
	s.mu.Unlock()
	s.sendMouseInput(0x0001, ix, iy, 0)
}
func (s *WindowsSimulator) Click(button string) {
	go func() {
		s.MouseDown(button)
		time.Sleep(12 * time.Millisecond)
		s.MouseUp(button)
	}()
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
		inputs := []rawInput{
			makeKeyboardInput(0x0004, 0, uint16(r)),
			makeKeyboardInput(0x0004|0x0002, 0, uint16(r)),
		}
		s.sendInputs(inputs)
	}
}
func (s *WindowsSimulator) Key(key string) {
	var vk uint16
	switch strings.ToLower(key) {
	case "backspace":
		vk = 0x08
	case "return", "enter":
		vk = 0x0D
	case "tab":
		vk = 0x09
	case "escape", "esc":
		vk = 0x1B
	case "space":
		vk = 0x20
	case "up":
		vk = 0x26
	case "down":
		vk = 0x28
	case "left":
		vk = 0x25
	case "right":
		vk = 0x27
	case "home":
		vk = 0x24
	case "end":
		vk = 0x23
	case "prior", "pageup":
		vk = 0x21
	case "next", "pagedown":
		vk = 0x22
	case "delete", "del":
		vk = 0x2E
	case "f1":
		vk = 0x70
	case "f2":
		vk = 0x71
	case "f3":
		vk = 0x72
	case "f4":
		vk = 0x73
	case "f5":
		vk = 0x74
	case "f6":
		vk = 0x75
	case "f7":
		vk = 0x76
	case "f8":
		vk = 0x77
	case "f9":
		vk = 0x78
	case "f10":
		vk = 0x79
	case "f11":
		vk = 0x7A
	case "f12":
		vk = 0x7B
	case "volumeup":
		vk = 0xAF
	case "volumedown":
		vk = 0xAE
	default:
		return
	}
	inputs := []rawInput{
		makeKeyboardInput(0, vk, 0),
		makeKeyboardInput(0x0002, vk, 0),
	}
	s.sendInputs(inputs)
}
func (s *WindowsSimulator) KeyCombo(modifier, key string) {
	var modVk uint16
	switch strings.ToLower(modifier) {
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
	if key == "+" || key == "=" {
		keyVk = 0xBB
	} else if key == "-" {
		keyVk = 0xBD
	} else if len(key) == 1 {
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
	inputs := []rawInput{
		makeKeyboardInput(0, modVk, 0),
		makeKeyboardInput(0, keyVk, 0),
		makeKeyboardInput(0x0002, keyVk, 0),
		makeKeyboardInput(0x0002, modVk, 0),
	}
	s.sendInputs(inputs)
}
func (s *WindowsSimulator) Close() {}
