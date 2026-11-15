package main

import (
	"strings"
	"testing"
)

// Issue #12: the terminal QR code must be plain black-on-white (no ANSI
// colour codes) and large enough to scan.
func TestRenderQRCodeIsPlainBlackOnWhite(t *testing.T) {
	// Wide terminal: full-block rendering, the easiest to scan.
	out := renderQRCodeSized("http://192.168.1.15:8080/?token=test&udp=8081", 120)
	if out == "" {
		t.Fatal("renderQRCode returned an empty string")
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("QR code must not contain ANSI colour codes, got:\n%s", out)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 30 {
		t.Fatalf("QR code looks too small: only %d lines, want at least 30", len(lines))
	}
	// Every line must consist only of black modules, spaces and the
	// quiet zone — nothing else.
	for _, line := range lines {
		for _, r := range line {
			if r != '█' && r != ' ' {
				t.Fatalf("unexpected character %q in QR output", r)
			}
		}
	}
	if !strings.Contains(out, "██") {
		t.Fatal("QR code contains no black modules")
	}
	// The quiet zone: the first and last lines must be entirely blank.
	if strings.TrimSpace(lines[0]) != "" || strings.TrimSpace(lines[len(lines)-1]) != "" {
		t.Fatal("QR code is missing a blank quiet zone at the top/bottom")
	}
}

// A narrow terminal gets the compact half-block rendering instead of
// an overflowing full-block one, still plain black-on-white.
func TestRenderQRCodeFitsNarrowTerminal(t *testing.T) {
	out := renderQRCodeSized("http://192.168.1.15:8080/?token=test&udp=8081", 60)
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("QR code must not contain ANSI colour codes, got:\n%s", out)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	for _, line := range lines {
		if len([]rune(line)) > 60 {
			t.Fatalf("QR line is %d columns wide, exceeds terminal width 60: %q", len([]rune(line)), line)
		}
		for _, r := range line {
			if r != '█' && r != '▀' && r != '▄' && r != ' ' {
				t.Fatalf("unexpected character %q in half-block QR output", r)
			}
		}
	}
	if !strings.Contains(out, "█") {
		t.Fatal("half-block QR code contains no black modules")
	}
}

// A terminal too narrow for any readable QR prints the URL as text
// instead of an overflowing block.
func TestRenderQRCodeTooNarrowFallsBackToURL(t *testing.T) {
	url := "http://192.168.1.15:8080/?token=test&udp=8081"
	out := renderQRCodeSized(url, 20)
	if !strings.Contains(out, url) {
		t.Fatalf("narrow-terminal fallback must print the URL, got:\n%s", out)
	}
}
