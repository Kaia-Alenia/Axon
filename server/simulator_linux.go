//go:build linux || darwin

package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

type LinuxSimulator struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	stdin    *bufio.Writer
	rawStdin io.WriteCloser
	ch       chan string
	closed   bool
}

func initSimulator() InputSimulator {
	s := &LinuxSimulator{
		ch: make(chan string, 100),
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

func (s *LinuxSimulator) loop() {
	for cmdStr := range s.ch {
		s.mu.Lock()
		cmd := s.cmd
		stdin := s.stdin
		s.mu.Unlock()

		if cmd == nil {
			newCmd := exec.Command("xdotool", "-")
			newRawStdin, err := newCmd.StdinPipe()
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if err := newCmd.Start(); err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			s.mu.Lock()
			s.cmd = newCmd
			s.rawStdin = newRawStdin
			s.stdin = bufio.NewWriter(newRawStdin)
			cmd = newCmd
			stdin = s.stdin
			s.mu.Unlock()
		}

		_, err := fmt.Fprintln(stdin, cmdStr)
		if err == nil {
			err = stdin.Flush()
		}
		if err != nil {
			s.mu.Lock()
			if s.cmd == cmd {
				_ = s.rawStdin.Close()
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
				s.cmd = nil
				s.stdin = nil
				s.rawStdin = nil
			}
			s.mu.Unlock()
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
	s.Send(fmt.Sprintf("key %s+%s", modifier, key))
}

func (s *LinuxSimulator) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	close(s.ch)
	
	cmd := s.cmd
	rawStdin := s.rawStdin
	s.cmd = nil
	s.stdin = nil
	s.rawStdin = nil
	s.mu.Unlock()

	if cmd != nil {
		_ = rawStdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
}
