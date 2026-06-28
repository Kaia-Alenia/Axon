package main
import (
	"strings"
	"testing"
	"unicode/utf8"
)
func TestGetRGBColor(t *testing.T) {
	tests := []struct {
		name      string
		phase     float64
		expectedR int
		expectedG int
		expectedB int
	}{
		{name: "Phase 0.0", phase: 0.0, expectedR: 127, expectedG: 236, expectedB: 17},
		{name: "Phase 0.25", phase: 0.25, expectedR: 254, expectedG: 63, expectedB: 63},
		{name: "Phase 0.333", phase: 0.3333333333333333, expectedR: 236, expectedG: 17, expectedB: 126},
		{name: "Phase 0.5", phase: 0.5, expectedR: 127, expectedG: 17, expectedB: 236},
		{name: "Phase 0.666", phase: 0.6666666666666666, expectedR: 17, expectedG: 126, expectedB: 236},
		{name: "Phase 0.75", phase: 0.75, expectedR: 0, expectedG: 190, expectedB: 190},
		{name: "Phase 1.0", phase: 1.0, expectedR: 126, expectedG: 236, expectedB: 17},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, g, b := getRGBColor(tc.phase)
			if r != tc.expectedR || g != tc.expectedG || b != tc.expectedB {
				t.Errorf("getRGBColor(%v) = (%v, %v, %v), want (%v, %v, %v)", tc.phase, r, g, b, tc.expectedR, tc.expectedG, tc.expectedB)
			}
		})
	}
}
var testQRStr = strings.Repeat("█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄█▀▄\n", 100)
func BenchmarkPrintQRRainbow_OriginalRunesCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lines := strings.Split(testQRStr, "\n")
		totalRunes := 0
		for _, line := range lines {
			totalRunes += len([]rune(line))
		}
		_ = totalRunes
	}
}
func BenchmarkPrintQRRainbow_OptimizedRunesCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lines := strings.Split(testQRStr, "\n")
		totalRunes := utf8.RuneCountInString(testQRStr) - len(lines) + 1
		_ = totalRunes
	}
}
func BenchmarkPrintQRRainbow_OriginalPrintLoop(b *testing.B) {
	lines := strings.Split(testQRStr, "\n")
	for i := 0; i < b.N; i++ {
		globalIdx := 0
		for _, line := range lines {
			for _, _ = range []rune(line) {
				globalIdx++
			}
		}
	}
}
func BenchmarkPrintQRRainbow_OptimizedPrintLoop(b *testing.B) {
	lines := strings.Split(testQRStr, "\n")
	for i := 0; i < b.N; i++ {
		globalIdx := 0
		for _, line := range lines {
			for _, _ = range line {
				globalIdx++
			}
		}
	}
}
func TestColorizeRainbow(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		frequency float64
		wantFunc  func(string) bool
	}{
		{
			name:      "Empty string",
			text:      "",
			frequency: 0.5,
			wantFunc: func(got string) bool {
				return got == "\033[0m"
			},
		},
		{
			name:      "Spaces and newlines only",
			text:      " \n \n",
			frequency: 0.5,
			wantFunc: func(got string) bool {
				return got == " \n \n\033[0m"
			},
		},
		{
			name:      "Normal string with zero frequency",
			text:      "a",
			frequency: 0.0,
			wantFunc: func(got string) bool {
				return strings.HasPrefix(got, "\033[38;2;") && strings.HasSuffix(got, "ma\033[0m")
			},
		},
		{
			name:      "Normal string with frequency",
			text:      "hello",
			frequency: 0.5,
			wantFunc: func(got string) bool {
				return strings.Contains(got, "\033[38;2;") &&
					strings.Contains(got, "h") &&
					strings.Contains(got, "e") &&
					strings.Contains(got, "l") &&
					strings.Contains(got, "o") &&
					strings.HasSuffix(got, "\033[0m") &&
					len(got) > len("hello")
			},
		},
		{
			name:      "Mixed string",
			text:      "a b\nc",
			frequency: 0.5,
			wantFunc: func(got string) bool {
				return strings.HasSuffix(got, "\033[0m") &&
					strings.Contains(got, " ") &&
					strings.Contains(got, "\n") &&
					strings.Contains(got, "b") &&
					strings.Contains(got, "c")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorizeRainbow(tt.text, tt.frequency)
			if !tt.wantFunc(got) {
				t.Errorf("colorizeRainbow() = %q, want matching wantFunc", got)
			}
		})
	}
}
