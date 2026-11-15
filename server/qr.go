package main

import (
	"fmt"
	"strings"

	"github.com/skip2/go-qrcode"
)

// printTerminalQRCode prints the connection URL as a QR code, sized to
// the console window and aware of its state.
//
// The window state matters: writing a huge block of text to a
// minimized console on Windows corrupts its buffer (conhost reflow)
// and crashes the server, so when the window is minimized nothing is
// printed here at all — the caller re-prints the banner and QR once
// the window is restored (see watchConsoleRestore). On a maximized or
// normal window the code is rendered to fit the measured terminal
// width instead of always at full size.
func printTerminalQRCode(url string) {
	if isConsoleMinimized() {
		return
	}
	fmt.Print(renderQRCode(url))
}

// renderQRCode renders url as a plain black-on-white QR code sized to
// the current terminal width. It never contains ANSI colour codes: a
// QR code must be high contrast to scan reliably.
func renderQRCode(url string) string {
	return renderQRCodeSized(url, terminalWidth())
}

// renderQRCodeSized renders url as a plain black-on-white QR code
// that fits in width terminal columns.
//
// Full-block rendering (each module two characters wide, "██", so the
// code stays square) is used whenever it fits — it is the easiest to
// scan. On a narrower window it falls back to half-block rendering
// (each cell holds two stacked modules: "▀" top only, "▄" bottom
// only, "█" both), which is half as wide and half as tall and still
// scans. On a terminal too narrow for even that, it prints the URL as
// text instead of an overflowing, unscannable block.
func renderQRCodeSized(url string, width int) string {
	qrCode, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return fmt.Sprintf("Error generating QR code in console: %v\n", err)
	}
	bitmap := qrCode.Bitmap()
	size := len(bitmap)
	if width <= 0 {
		width = 80
	}
	const quietZone = 4
	if fullWidth := (size + 2*quietZone) * 2; fullWidth <= width {
		return renderFullBlock(bitmap, quietZone)
	}
	const halfQuiet = 2
	if halfWidth := size + 2*halfQuiet; halfWidth <= width {
		return renderHalfBlock(bitmap, halfQuiet)
	}
	return fmt.Sprintf("QR code does not fit this terminal (width %d columns).\nEnter the address manually: %s\n", width, url)
}

// renderFullBlock renders bitmap with one module per two characters
// ("██" black, two spaces white) and a quietZone-module white border.
func renderFullBlock(bitmap [][]bool, quietZone int) string {
	size := len(bitmap)
	var sb strings.Builder
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

// paddedModule reads bitmap at padded coordinates (r, c), where the
// bitmap is centred in a quiet-module white border. Out-of-range
// coordinates are white.
func paddedModule(bitmap [][]bool, quiet, r, c int) bool {
	size := len(bitmap)
	rr, cc := r-quiet, c-quiet
	if rr < 0 || rr >= size || cc < 0 || cc >= size {
		return false
	}
	return bitmap[rr][cc]
}

// renderHalfBlock renders bitmap with one character per module column
// and two module rows per line, using half-block characters. quiet is
// the white border in modules on every side.
func renderHalfBlock(bitmap [][]bool, quiet int) string {
	size := len(bitmap)
	total := size + 2*quiet
	var sb strings.Builder
	for line := 0; line*2 < total; line++ {
		topRow, bottomRow := line*2, line*2+1
		for col := 0; col < total; col++ {
			top := paddedModule(bitmap, quiet, topRow, col)
			bottom := false
			if bottomRow < total {
				bottom = paddedModule(bitmap, quiet, bottomRow, col)
			}
			switch {
			case top && bottom:
				sb.WriteString("█")
			case top:
				sb.WriteString("▀")
			case bottom:
				sb.WriteString("▄")
			default:
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
