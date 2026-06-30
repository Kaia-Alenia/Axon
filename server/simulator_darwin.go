//go:build darwin

package main

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

type DarwinSimulator struct {
	mu     sync.Mutex
	ch     chan string
	closed bool
}

func initSimulator() InputSimulator {
	s := &DarwinSimulator{
		ch: make(chan string, 200),
	}
	if _, err := exec.LookPath("cliclick"); err != nil {
		fmt.Println("[INPUT] Warning: cliclick is not installed. Mouse and keyboard simulation will not work.")
		fmt.Println("[INPUT] Install it with: brew install cliclick")
	}
	go s.loop()
	return s
}

func (s *DarwinSimulator) Send(cmd string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	select {
	case s.ch <- cmd:
	default:
	}
}

func (s *DarwinSimulator) loop() {
	for cmd := range s.ch {
		s.execute(cmd)
	}
}

func (s *DarwinSimulator) execute(cmd string) {
	parts := strings.SplitN(cmd, " ", 2)
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "move":
		if len(parts) == 2 {
			_ = exec.Command("cliclick", "m:"+parts[1]).Run()
		}
	case "click":
		_ = exec.Command("cliclick", "c:.").Run()
	case "rclick":
		_ = exec.Command("cliclick", "rc:.").Run()
	case "mdown":
		_ = exec.Command("cliclick", "dd:.").Run()
	case "mup":
		_ = exec.Command("cliclick", "du:.").Run()
	case "scrollup":
		_ = exec.Command("cliclick", "ku:arrow-up").Run()
	case "scrolldown":
		_ = exec.Command("cliclick", "kd:arrow-down").Run()
	case "type":
		if len(parts) == 2 {
			_ = exec.Command("cliclick", "t:"+parts[1]).Run()
		}
	case "key":
		if len(parts) == 2 {
			_ = exec.Command("cliclick", "kp:"+parts[1]).Run()
		}
	}
}

func (s *DarwinSimulator) MoveMouse(dx, dy float64) {
	s.Send(fmt.Sprintf("move %d,%d", int(dx), int(dy)))
}

func (s *DarwinSimulator) Click(button string) {
	if button == "right" {
		s.Send("rclick")
	} else {
		s.Send("click")
	}
}

func (s *DarwinSimulator) MouseDown(button string) {
	s.Send("mdown")
}

func (s *DarwinSimulator) MouseUp(button string) {
	s.Send("mup")
}

func (s *DarwinSimulator) Scroll(direction string) {
	if direction == "up" {
		s.Send("scrollup")
	} else {
		s.Send("scrolldown")
	}
}

func (s *DarwinSimulator) Type(text string) {
	s.Send("type " + text)
}

func (s *DarwinSimulator) Key(key string) {
	darwinKey := key
	switch key {
	case "Return":
		darwinKey = "return"
	case "BackSpace":
		darwinKey = "delete"
	case "Escape":
		darwinKey = "escape"
	case "Tab":
		darwinKey = "tab"
	case "Up":
		darwinKey = "arrow-up"
	case "Down":
		darwinKey = "arrow-down"
	case "Left":
		darwinKey = "arrow-left"
	case "Right":
		darwinKey = "arrow-right"
	case "Prior":
		darwinKey = "page-up"
	case "Next":
		darwinKey = "page-down"
	case "Delete":
		darwinKey = "forward-delete"
	case "Home":
		darwinKey = "home"
	case "End":
		darwinKey = "end"
	}
	s.Send("key " + darwinKey)
}

func (s *DarwinSimulator) KeyCombo(modifier, key string) {
	mod := ""
	switch strings.ToLower(modifier) {
	case "ctrl":
		mod = "cmd"
	case "alt":
		mod = "alt"
	case "shift":
		mod = "shift"
	}
	darwinKey := strings.ToLower(key)
	if mod != "" {
		_ = exec.Command("cliclick", fmt.Sprintf("kd:%s", mod), fmt.Sprintf("kp:%s", darwinKey), fmt.Sprintf("ku:%s", mod)).Run()
	} else {
		s.Send("key " + darwinKey)
	}
}

func (s *DarwinSimulator) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.ch)
}
