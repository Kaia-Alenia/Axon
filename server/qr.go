package main

import (
	"fmt"
	"strings"

	"github.com/skip2/go-qrcode"
)

// printTerminalQRCode prints the connection URL as a large, plain
// black-on-white QR code.
//
// The previous implementation rendered the QR with a rainbow ANSI colour
// gradient and go-qrcode's small half-block rendering, which made it hard
// to scan: the coloured modules lowered contrast and the code was too
// small on screen (issue #12). A QR code must be high contrast to scan
// reliably, so this renderer deliberately uses no colour at all: black
// modules on the terminal's default (white/light) background, each module
// drawn two characters wide, with a 4-module quiet zone around it.
func printTerminalQRCode(url string) {
	fmt.Print(renderQRCode(url))
}

// renderQRCode renders url as a plain black-on-white QR code string.
// Each module is two characters wide ("██" for black, two spaces for
// white) so the code stays square, and a 4-module white quiet zone is
// added on every side as the QR spec requires.
func renderQRCode(url string) string {
	qrCode, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return fmt.Sprintf("Error generating QR code in console: %v\n", err)
	}
	const quietZone = 4
	bitmap := qrCode.Bitmap()
	size := len(bitmap)
	var sb strings.Builder
	// Top quiet zone.
	blankLine := strings.Repeat("  ", size+2*quietZone) + "\n"
	for i := 0; i < quietZone; i++ {
		sb.WriteString(blankLine)
	}
	for _, row := range bitmap {
		sb.WriteString(strings.Repeat("  ", quietZone))
		for _, module := range row {
			if module {
				sb.WriteString("██")
			} else {
				sb.WriteString("  ")
			}
		}
		sb.WriteString(strings.Repeat("  ", quietZone))
		sb.WriteString("\n")
	}
	for i := 0; i < quietZone; i++ {
		sb.WriteString(blankLine)
	}
	return sb.String()
}
