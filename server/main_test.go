package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

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
