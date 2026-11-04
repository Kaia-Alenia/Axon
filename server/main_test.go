package main

import (
	"io"
	"net/http"
	"net/http/httptest"
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
			diffB := b - tc.expectedB
			if diffB < 0 {
				diffB = -diffB
			}
			if r != tc.expectedR || g != tc.expectedG || diffB > 1 {
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

func TestGenerateToken(t *testing.T) {
	token1 := generateToken()
	token2 := generateToken()

	if len(token1) != 32 {
		t.Errorf("Expected token length 32, got %d", len(token1))
	}
	if len(token2) != 32 {
		t.Errorf("Expected token length 32, got %d", len(token2))
	}

	if token1 == token2 {
		t.Errorf("Expected tokens to be unique, got %s twice", token1)
	}

	for _, char := range token1 {
		if !strings.ContainsRune("0123456789abcdef", char) {
			t.Errorf("Expected hex token, got non-hex char %c", char)
		}
	}
}

func TestHandleAuthentication(t *testing.T) {
	activeToken = "test-token-123"

	tests := []struct {
		name           string
		token          string
		expectedResult bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid Token",
			token:          "test-token-123",
			expectedResult: true,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Invalid Token",
			token:          "wrong-token",
			expectedResult: false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
		{
			name:           "Missing Token",
			token:          "",
			expectedResult: false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Unauthorized\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?token="+tc.token, nil)
			if tc.token == "" {
				req = httptest.NewRequest(http.MethodGet, "/", nil)
			}

			w := httptest.NewRecorder()

			result := handleAuthentication(w, req)

			if result != tc.expectedResult {
				t.Errorf("handleAuthentication() returned %v, expected %v", result, tc.expectedResult)
			}

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, res.StatusCode)
			}

			if !tc.expectedResult {
				body, _ := io.ReadAll(res.Body)
				if string(body) != tc.expectedBody {
					t.Errorf("Expected body %q, got %q", tc.expectedBody, string(body))
				}
			}
		})
	}
}

func TestIsVirtualInterface(t *testing.T) {
	if !isVirtualInterface("vEthernet (WSL)") {
		t.Errorf("Expected vEthernet (WSL) to be detected as virtual")
	}
	if !isVirtualInterface("VirtualBox Host-Only Network") {
		t.Errorf("Expected VirtualBox Host-Only Network to be detected as virtual")
	}
	if isVirtualInterface("Wi-Fi") {
		t.Errorf("Expected Wi-Fi not to be detected as virtual")
	}
	if isVirtualInterface("Ethernet") {
		t.Errorf("Expected Ethernet not to be detected as virtual")
	}
}
