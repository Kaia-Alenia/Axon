package main

import (
	"strings"
	"testing"
)

// Issue #12: the terminal QR code must be plain black-on-white (no ANSI
// colour codes) and large enough to scan.
func TestRenderQRCodeIsPlainBlackOnWhite(t *testing.T) {
	out := renderQRCode("http://192.168.1.15:8080/?token=test&udp=8081")
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
