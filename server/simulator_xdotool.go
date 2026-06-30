//go:build linux

package main

import (
	"fmt"
	"github.com/bendahl/uinput"
	"os/exec"
	"strings"
	"sync"
)

var keyMap = map[string]int{
	"Return":     uinput.KeyEnter,
	"BackSpace":  uinput.KeyBackspace,
	"space":      uinput.KeySpace,
	"Escape":     uinput.KeyEsc,
	"Tab":        uinput.KeyTab,
	"Up":         uinput.KeyUp,
	"Down":       uinput.KeyDown,
	"Left":       uinput.KeyLeft,
	"Right":      uinput.KeyRight,
	"Home":       uinput.KeyHome,
	"End":        uinput.KeyEnd,
	"Prior":      uinput.KeyPageup,
	"Next":       uinput.KeyPagedown,
	"Delete":     uinput.KeyDelete,
	"F5":         uinput.KeyF5,
	"volumeup":   uinput.KeyVolumeup,
	"volumedown": uinput.KeyVolumedown,
}

type LinuxSimulator struct {
	mu       sync.Mutex
	mouse    uinput.Mouse
	keyboard uinput.Keyboard
	ch       chan string
	closed   bool
}

func initSimulator() InputSimulator {
	s := &LinuxSimulator{
		ch: make(chan string, 200),
	}
	var err error
	s.mouse, err = uinput.CreateMouse("/dev/uinput", []byte("axon-mouse"))
	if err != nil {
		fmt.Println("Warning: uinput mouse failed, using xdotool fallback:", err)
	}
	s.keyboard, err = uinput.CreateKeyboard("/dev/uinput", []byte("axon-keyboard"))
	if err != nil {
		fmt.Println("Warning: uinput keyboard failed, using xdotool fallback:", err)
	}
	if s.mouse == nil || s.keyboard == nil {
		if _, err := exec.LookPath("xdotool"); err != nil {
			fmt.Println("Warning: xdotool is not installed. Fallback input simulation will fail.")
		}
	}
	go s.loop()
	return s
}
func (s *LinuxSimulator) Send(cmdStr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	select {
	case s.ch <- cmdStr:
	default:
	}
}
func (s *LinuxSimulator) executeCmd(cmdStr string) {
	if strings.HasPrefix(cmdStr, "mousemove_relative") {
		var dx, dy float64
		_, err := fmt.Sscanf(cmdStr, "mousemove_relative -- %f %f", &dx, &dy)
		if err == nil {
			if s.mouse != nil {
				_ = s.mouse.Move(int32(dx), int32(dy))
			} else {
				_ = exec.Command("xdotool", "mousemove_relative", "--", fmt.Sprintf("%d", int32(dx)), fmt.Sprintf("%d", int32(dy))).Run()
			}
		}
	} else if strings.HasPrefix(cmdStr, "click") {
		var btn string
		var count int
		_, err := fmt.Sscanf(cmdStr, "click --repeat %d %s", &count, &btn)
		if err != nil {
			_, err = fmt.Sscanf(cmdStr, "click %s", &btn)
			count = 1
		}
		if err == nil {
			if s.mouse != nil {
				for i := 0; i < count; i++ {
					switch btn {
					case "1":
						_ = s.mouse.LeftClick()
					case "2":
						_ = s.mouse.MiddleClick()
					case "3":
						_ = s.mouse.RightClick()
					case "4":
						_ = s.mouse.Wheel(false, 1)
					case "5":
						_ = s.mouse.Wheel(false, -1)
					}
				}
			} else {
				args := []string{"click"}
				if count > 1 {
					args = append(args, "--repeat", fmt.Sprintf("%d", count))
				}
				args = append(args, btn)
				_ = exec.Command("xdotool", args...).Run()
			}
		}
	} else if strings.HasPrefix(cmdStr, "mousedown") {
		var btn string
		_, err := fmt.Sscanf(cmdStr, "mousedown %s", &btn)
		if err == nil {
			if s.mouse != nil {
				switch btn {
				case "1":
					_ = s.mouse.LeftPress()
				case "3":
					_ = s.mouse.RightPress()
				}
			} else {
				_ = exec.Command("xdotool", "mousedown", btn).Run()
			}
		}
	} else if strings.HasPrefix(cmdStr, "mouseup") {
		var btn string
		_, err := fmt.Sscanf(cmdStr, "mouseup %s", &btn)
		if err == nil {
			if s.mouse != nil {
				switch btn {
				case "1":
					_ = s.mouse.LeftRelease()
				case "3":
					_ = s.mouse.RightRelease()
				}
			} else {
				_ = exec.Command("xdotool", "mouseup", btn).Run()
			}
		}
	} else if strings.HasPrefix(cmdStr, "type") {
		parts := strings.SplitN(cmdStr, " -- ", 2)
		if len(parts) == 2 {
			text := parts[1]
			if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
				text = text[1 : len(text)-1]
			}
			if s.keyboard != nil {
				for _, char := range text {
					s.typeChar(char)
				}
			} else {
				_ = exec.Command("xdotool", "type", "--", text).Run()
			}
		}
	} else if strings.HasPrefix(cmdStr, "key ") {
		key := strings.TrimPrefix(cmdStr, "key ")
		if strings.Contains(key, "+") {
			parts := strings.Split(key, "+")
			if len(parts) == 2 {
				s.KeyCombo(parts[0], parts[1])
			}
		} else {
			if s.keyboard != nil {
				if code, ok := keyMap[key]; ok {
					_ = s.keyboard.KeyPress(code)
				} else if len(key) == 1 {
					s.typeChar(rune(key[0]))
				}
			} else {
				xKey := key
				switch key {
				case "Prior":
					xKey = "Page_Up"
				case "Next":
					xKey = "Page_Down"
				case "volumeup":
					xKey = "XF86AudioRaiseVolume"
				case "volumedown":
					xKey = "XF86AudioLowerVolume"
				}
				_ = exec.Command("xdotool", "key", xKey).Run()
			}
		}
	}
}

var runeToKey = map[rune]int{
	'a': uinput.KeyA, 'b': uinput.KeyB, 'c': uinput.KeyC, 'd': uinput.KeyD,
	'e': uinput.KeyE, 'f': uinput.KeyF, 'g': uinput.KeyG, 'h': uinput.KeyH,
	'i': uinput.KeyI, 'j': uinput.KeyJ, 'k': uinput.KeyK, 'l': uinput.KeyL,
	'm': uinput.KeyM, 'n': uinput.KeyN, 'o': uinput.KeyO, 'p': uinput.KeyP,
	'q': uinput.KeyQ, 'r': uinput.KeyR, 's': uinput.KeyS, 't': uinput.KeyT,
	'u': uinput.KeyU, 'v': uinput.KeyV, 'w': uinput.KeyW, 'x': uinput.KeyX,
	'y': uinput.KeyY, 'z': uinput.KeyZ,
	'0': uinput.Key0, '1': uinput.Key1, '2': uinput.Key2, '3': uinput.Key3,
	'4': uinput.Key4, '5': uinput.Key5, '6': uinput.Key6, '7': uinput.Key7,
	'8': uinput.Key8, '9': uinput.Key9,
	' ': uinput.KeySpace, '\n': uinput.KeyEnter,
	',': uinput.KeyComma, '.': uinput.KeyDot,
	'-': uinput.KeySlash,
	'¿': uinput.KeyEqual,
	'´': uinput.KeyLeftbrace,
	'+': uinput.KeyRightbrace,
	'{': uinput.KeyApostrophe,
	'}': uinput.KeyBackslash,
	'<': 86,
}
var shiftRuneMap = map[rune]rune{
	'!': '1', '"': '2', '#': '3', '$': '4', '%': '5',
	'&': '6', '/': '7', '(': '8', ')': '9', '=': '0',
	'?': '\'',
	'¡': '¿',
	'¨': '´',
	'*': '+',
	'[': '{',
	']': '}',
	'_': '-',
	'>': '<',
	';': ',',
	':': '.',
}

func (s *LinuxSimulator) typeChar(char rune) {
	if s.keyboard == nil {
		return
	}
	accentMap := map[rune]struct {
		vowelCode int
		shifted   bool
	}{
		'á': {uinput.KeyA, false},
		'é': {uinput.KeyE, false},
		'í': {uinput.KeyI, false},
		'ó': {uinput.KeyO, false},
		'ú': {uinput.KeyU, false},
		'Á': {uinput.KeyA, true},
		'É': {uinput.KeyE, true},
		'Í': {uinput.KeyI, true},
		'Ó': {uinput.KeyO, true},
		'Ú': {uinput.KeyU, true},
	}
	if char == 'ñ' {
		_ = s.keyboard.KeyPress(uinput.KeySemicolon)
	} else if char == 'Ñ' {
		_ = s.keyboard.KeyDown(uinput.KeyLeftshift)
		_ = s.keyboard.KeyPress(uinput.KeySemicolon)
		_ = s.keyboard.KeyUp(uinput.KeyLeftshift)
	} else if info, ok := accentMap[char]; ok {
		_ = s.keyboard.KeyPress(uinput.KeyLeftbrace)
		if info.shifted {
			_ = s.keyboard.KeyDown(uinput.KeyLeftshift)
			_ = s.keyboard.KeyPress(info.vowelCode)
			_ = s.keyboard.KeyUp(uinput.KeyLeftshift)
		} else {
			_ = s.keyboard.KeyPress(info.vowelCode)
		}
	} else if char >= 'a' && char <= 'z' {
		if code, ok := runeToKey[char]; ok {
			_ = s.keyboard.KeyPress(code)
		}
	} else if char >= 'A' && char <= 'Z' {
		lower := char - 'A' + 'a'
		if code, ok := runeToKey[lower]; ok {
			_ = s.keyboard.KeyDown(uinput.KeyLeftshift)
			_ = s.keyboard.KeyPress(code)
			_ = s.keyboard.KeyUp(uinput.KeyLeftshift)
		}
	} else if baseChar, isShifted := shiftRuneMap[char]; isShifted {
		if code, ok := runeToKey[baseChar]; ok {
			_ = s.keyboard.KeyDown(uinput.KeyLeftshift)
			_ = s.keyboard.KeyPress(code)
			_ = s.keyboard.KeyUp(uinput.KeyLeftshift)
		}
	} else {
		if code, ok := runeToKey[char]; ok {
			_ = s.keyboard.KeyPress(code)
		}
	}
}
func (s *LinuxSimulator) loop() {
	for cmdStr := range s.ch {
		if strings.HasPrefix(cmdStr, "mousemove_relative") {
			var dx, dy float64
			_, _ = fmt.Sscanf(cmdStr, "mousemove_relative -- %f %f", &dx, &dy)
			for {
				var nextCmd string
				var ok bool
				select {
				case nextCmd, ok = <-s.ch:
				default:
				}
				if !ok || nextCmd == "" {
					break
				}
				if strings.HasPrefix(nextCmd, "mousemove_relative") {
					var ndx, ndy float64
					_, _ = fmt.Sscanf(nextCmd, "mousemove_relative -- %f %f", &ndx, &ndy)
					dx += ndx
					dy += ndy
				} else {
					if dx != 0 || dy != 0 {
						s.executeCmd(fmt.Sprintf("mousemove_relative -- %f %f", dx, dy))
						dx = 0
						dy = 0
					}
					s.executeCmd(nextCmd)
				}
			}
			if dx != 0 || dy != 0 {
				s.executeCmd(fmt.Sprintf("mousemove_relative -- %f %f", dx, dy))
			}
		} else if strings.HasPrefix(cmdStr, "click 4") || strings.HasPrefix(cmdStr, "click 5") {
			btn := "4"
			if strings.HasPrefix(cmdStr, "click 5") {
				btn = "5"
			}
			count := 1
			for {
				var nextCmd string
				var ok bool
				select {
				case nextCmd, ok = <-s.ch:
				default:
				}
				if !ok || nextCmd == "" {
					break
				}
				if nextCmd == cmdStr {
					count++
				} else {
					if count > 1 {
						s.executeCmd(fmt.Sprintf("click --repeat %d %s", count, btn))
					} else {
						s.executeCmd(cmdStr)
					}
					s.executeCmd(nextCmd)
					count = 0
					break
				}
			}
			if count > 0 {
				if count > 1 {
					s.executeCmd(fmt.Sprintf("click --repeat %d %s", count, btn))
				} else {
					s.executeCmd(cmdStr)
				}
			}
		} else {
			s.executeCmd(cmdStr)
		}
	}
}
func (s *LinuxSimulator) MoveMouse(dx, dy float64) {
	s.Send(fmt.Sprintf("mousemove_relative -- %f %f", dx, dy))
}
func (s *LinuxSimulator) Click(button string) {
	btn := "1"
	if button == "right" {
		btn = "3"
	} else if button == "middle" {
		btn = "2"
	}
	s.Send(fmt.Sprintf("click %s", btn))
}
func (s *LinuxSimulator) MouseDown(button string) {
	btn := "1"
	if button == "right" {
		btn = "3"
	}
	s.Send(fmt.Sprintf("mousedown %s", btn))
}
func (s *LinuxSimulator) MouseUp(button string) {
	btn := "1"
	if button == "right" {
		btn = "3"
	}
	s.Send(fmt.Sprintf("mouseup %s", btn))
}
func (s *LinuxSimulator) Scroll(direction string) {
	if direction == "up" {
		s.Send("click 4")
	} else {
		s.Send("click 5")
	}
}
func (s *LinuxSimulator) Type(text string) {
	s.Send(fmt.Sprintf("type -- %s", text))
}
func (s *LinuxSimulator) Key(key string) {
	s.Send(fmt.Sprintf("key %s", key))
}
func (s *LinuxSimulator) KeyCombo(modifier, key string) {
	if s.keyboard != nil {
		modCode := 0
		switch strings.ToLower(modifier) {
		case "ctrl":
			modCode = uinput.KeyLeftctrl
		case "alt":
			modCode = uinput.KeyLeftalt
		case "shift":
			modCode = uinput.KeyLeftshift
		}
		var keyCode int
		if len(key) == 1 {
			char := rune(strings.ToLower(key)[0])
			keyCode = runeToKey[char]
		} else {
			keyCode = keyMap[key]
		}
		if keyCode != 0 {
			if modCode != 0 {
				_ = s.keyboard.KeyDown(modCode)
			}
			_ = s.keyboard.KeyPress(keyCode)
			if modCode != 0 {
				_ = s.keyboard.KeyUp(modCode)
			}
		}
	} else {
		xKey := key
		switch key {
		case "Prior":
			xKey = "Page_Up"
		case "Next":
			xKey = "Page_Down"
		}
		combo := fmt.Sprintf("%s+%s", strings.ToLower(modifier), xKey)
		_ = exec.Command("xdotool", "key", combo).Run()
	}
}
func (s *LinuxSimulator) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	close(s.ch)
	mouse := s.mouse
	keyboard := s.keyboard
	s.mouse = nil
	s.keyboard = nil
	s.mu.Unlock()
	if mouse != nil {
		_ = mouse.Close()
	}
	if keyboard != nil {
		_ = keyboard.Close()
	}
}
