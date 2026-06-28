package main
import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
	"github.com/skip2/go-qrcode"
)
func printTerminalQRCode(url string) {
	qrCode, err := qrcode.New(url, qrcode.Medium)
	if err == nil {
		printQRRainbow(qrCode.ToSmallString(false))
	} else {
		fmt.Println("Error generating QR code in console:", err)
	}
}
func printQRRainbow(qrStr string) {
	lines := strings.Split(qrStr, "\n")
	totalRunes := utf8.RuneCountInString(qrStr) - len(lines) + 1
	globalIdx := 0
	frequency := 1.2
	var sb strings.Builder
	sb.Grow(len(qrStr) * 19)
	var buf [16]byte
	for _, line := range lines {
		for _, ch := range line {
			phase := float64(globalIdx) / float64(totalRunes+1) * frequency
			r, g, b := getRGBColor(phase)
			sb.WriteString("\033[38;2;")
			b1 := strconv.AppendInt(buf[:0], int64(r), 10)
			b1 = append(b1, ';')
			b1 = strconv.AppendInt(b1, int64(g), 10)
			b1 = append(b1, ';')
			b1 = strconv.AppendInt(b1, int64(b), 10)
			b1 = append(b1, 'm')
			sb.Write(b1)
			sb.WriteRune(ch)
			globalIdx++
		}
		sb.WriteString("\033[0m\n")
	}
	fmt.Print(sb.String())
}
