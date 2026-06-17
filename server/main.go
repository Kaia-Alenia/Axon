package main

import (
	"crypto/rand"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
)

//go:embed web/*
var webFS embed.FS

func setupADBReverse() {
	_ = exec.Command("adb", "reverse", "tcp:6969", "tcp:6969").Run()
}

func getRGBColor(phase float64) (int, int, int) {
	r := int(127.0 + 127.0*math.Sin(2.0*math.Pi*phase+0.0))
	g := int(127.0 + 127.0*math.Sin(2.0*math.Pi*phase+2.0*math.Pi/3.0))
	b := int(127.0 + 127.0*math.Sin(2.0*math.Pi*phase+4.0*math.Pi/3.0))
	return r, g, b
}

func colorizeRainbow(text string, frequency float64) string {
	var result strings.Builder
	runes := []rune(text)
	length := len(runes)
	for i, char := range runes {
		if char == ' ' || char == '\n' {
			result.WriteRune(char)
			continue
		}
		phase := float64(i) / float64(length) * frequency
		r, g, b := getRGBColor(phase)
		result.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm%c", r, g, b, char))
	}
	result.WriteString("\033[0m")
	return result.String()
}

func printLine(text string, delay time.Duration) {
	fmt.Println(text)
	time.Sleep(delay)
}

func getLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	priorityPrefixes := []string{"wlan", "wlp", "wlo", "en", "eth"}

	for _, prefix := range priorityPrefixes {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			if !strings.HasPrefix(iface.Name, prefix) {
				continue
			}
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

var (
	activeClientIP string
	activeClientMu sync.RWMutex
	activeToken    string
	simulator      InputSimulator
)

type ClientMessage struct {
	Type      string  `json:"type"`
	Dx        float64 `json:"dx"`
	Dy        float64 `json:"dy"`
	Button    string  `json:"button"`
	Text      string  `json:"text"`
	Key       string  `json:"key"`
	Modifier  string  `json:"modifier"`
	Timestamp int64   `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func generateToken() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token != activeToken {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	if tcpConn, ok := conn.UnderlyingConn().(*net.TCPConn); ok {
		_ = tcpConn.SetNoDelay(true)
	}

	remoteAddrStr := conn.RemoteAddr().String()
	ip, _, _ := net.SplitHostPort(remoteAddrStr)

	activeClientMu.Lock()
	activeClientIP = ip
	activeClientMu.Unlock()

	fmt.Printf("[CONEXIÓN] Cliente conectado desde: %s\n", remoteAddrStr)

	done := make(chan bool)
	defer func() {
		close(done)
		activeClientMu.Lock()
		if activeClientIP == ip {
			activeClientIP = ""
		}
		activeClientMu.Unlock()
		fmt.Println("[DESCONEXIÓN] Cliente desconectado")
	}()

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				pingMsg := map[string]interface{}{
					"type":      "ping",
					"timestamp": time.Now().UnixMilli(),
				}
				data, err := json.Marshal(pingMsg)
				if err != nil {
					continue
				}
				_ = conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
				err = conn.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg ClientMessage
		err = json.Unmarshal(message, &msg)
		if err != nil {
			continue
		}

		switch msg.Type {
		case "pong":
			rtt := time.Now().UnixMilli() - msg.Timestamp
			fmt.Printf("[LATENCIA] Latencia RTT: %d ms\n", rtt)
		case "move":
			simulator.MoveMouse(msg.Dx, msg.Dy)
		case "click":
			simulator.Click(msg.Button)
		case "mousedown":
			simulator.MouseDown(msg.Button)
		case "mouseup":
			simulator.MouseUp(msg.Button)
		case "scroll":
			dir := "down"
			if msg.Dy > 0 {
				dir = "up"
			}
			simulator.Scroll(dir)
		case "type":
			fmt.Printf("[TECLADO] Escribiendo texto: %q\n", msg.Text)
			simulator.Type(msg.Text)
		case "key":
			fmt.Printf("[TECLADO] Presionando tecla especial: %q\n", msg.Key)
			simulator.Key(msg.Key)
		case "keycombo":
			fmt.Printf("[TECLADO] Presionando combo: %s+%s\n", msg.Modifier, msg.Key)
			simulator.KeyCombo(msg.Modifier, msg.Key)
		}
	}
}

func startUDPServer() {
	addr, err := net.ResolveUDPAddr("udp", ":6970")
	if err != nil {
		fmt.Println("Error al resolver dirección UDP:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error al iniciar servidor UDP:", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 128)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		activeClientMu.RLock()
		clientIP := activeClientIP
		activeClientMu.RUnlock()

		if clientIP == "" || remoteAddr.IP.String() != clientIP {
			continue
		}

		payload := string(buf[:n])
		parts := strings.Split(payload, ":")
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "m":
			if len(parts) == 3 {
				dx, err1 := strconv.ParseFloat(parts[1], 64)
				dy, err2 := strconv.ParseFloat(parts[2], 64)
				if err1 == nil && err2 == nil {
					simulator.MoveMouse(dx, dy)
				}
			}
		case "s":
			if len(parts) == 2 {
				dy, err1 := strconv.ParseFloat(parts[1], 64)
				if err1 == nil {
					dir := "down"
					if dy > 0 {
						dir = "up"
					}
					simulator.Scroll(dir)
				}
			}
		}
	}
}

func printQRRainbow(qrStr string) {
	lines := strings.Split(qrStr, "\n")
	totalRunes := 0
	for _, line := range lines {
		totalRunes += len([]rune(line))
	}

	globalIdx := 0
	frequency := 1.2

	for _, line := range lines {
		var sb strings.Builder
		for _, ch := range []rune(line) {
			phase := float64(globalIdx) / float64(totalRunes+1) * frequency
			r, g, b := getRGBColor(phase)
			sb.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm%c", r, g, b, ch))
			globalIdx++
		}
		sb.WriteString("\033[0m")
		fmt.Println(sb.String())
		time.Sleep(18 * time.Millisecond)
	}
}

func main() {
	simulator = NewSimulator()
	defer simulator.Close()

	setupADBReverse()

	ip := getLocalIP()
	if ip == "" {
		fmt.Println("Error: No se pudo obtener la IP local")
		os.Exit(1)
	}

	activeToken = generateToken()
	addr := ip + ":6969"
	url := fmt.Sprintf("http://%s/?token=%s", addr, activeToken)

	lineDelay := 55 * time.Millisecond

	printLine(colorizeRainbow("    ___  _  ______  _  __", 0.4), lineDelay)
	printLine(colorizeRainbow("   /   || |/ / __ \\/ |/ /", 0.4), lineDelay)
	printLine(colorizeRainbow("  / /| ||   / / / /    / ", 0.4), lineDelay)
	printLine(colorizeRainbow(" / ___ |/   / /_/ /    /  ", 0.4), lineDelay)
	printLine(colorizeRainbow("/_/  |_/_/|_\\____/_/|_/   ", 0.4), lineDelay)
	printLine(colorizeRainbow("    [ HIGH-SPEED INPUT BRIDGE ]", 0.6), lineDelay)
	printLine("", 0)

	printLine(colorizeRainbow("      ⢀⣀⣀⣀⡀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⣀⣀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀", 0.8), lineDelay)
	printLine(colorizeRainbow("  ⣠⡾⠋⠉⠉⠉⠻⣷⠶⠞⠛⠛⠛⠶⠾⠋⠉⠉⠉⠻⢷⡄⠀⠀⠀⠀⠀⠀⠀⢀⣴⢟⣿⡟⠀⣤⡶⣦", 0.8), lineDelay)
	printLine(colorizeRainbow(" ⢠⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⣿⠀⠀⠀⠀⠀⠀⣴⡿⠃⣼⡿⠀⣾⡟⣼⠇⠀⠀⠀⠀⠀⠀⠀⣿⣿⠇", 0.8), lineDelay)
	printLine(colorizeRainbow("  ⢿⣄⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣀⠀⠀⢀⣴⡟⠀⢀⣤⡶⠶⣾⣿⣦⣤⣿⣃⣸⣿⡿⢃⣤⢶⣄⢠⣄⣤⣤⠀⣨⣧⠀⣠⣤⣦⣤", 0.8), lineDelay)
	printLine(colorizeRainbow("  ⢰⡟⠁⠀⣤⣤⣤⣤⡀⠀⠀⠀⢀⣶⣿⣿⣿⣿⣆⠈⢻⡆⢀⠹⠇⣠⣾⡟⠀⠀⢹⣿⠉⣿⡏⠀⣿⣧⠞⣩⣿⡟⣹⡟⢠⣿⠇⣼⡿⢁⣾⡟⢀⡄", 0.8), lineDelay)
	printLine(colorizeRainbow("  ⢻⡇⠀⠐⣟⣉⡉⠙⠓⠀⢀⡀⠘⣿⣿⣿⣿⣼⠃⠀⢸⡗⢿⣷⣾⡿⠋⠀⠀⠀⠘⢿⡶⠛⠿⠟⠻⠷⠞⠛⠿⠁⠻⠿⠛⠿⠞⠻⠿⠛⠻⠷⠛", 0.8), lineDelay)
	printLine(colorizeRainbow("  ⠈⢷⣄⠘⠛⠛⠀⠀⠸⣷⠿⠻⣾⠏⠈⠉⠽⠿⢟⣠⡿⠁⠀⠉⠉⠀⠀⠀⣶⡆⠰⣶⠆⡆⢰⠀⡶⢦⢠⡆⢠⠶⣄⢰⣶", 0.8), lineDelay)
	printLine(colorizeRainbow("    ⠉⠛⠷⠶⠶⠶⠆⠙⠓⠚⠋⠰⠶⠶⠶⠞⠛⠉⠀⠀⠀⠀⠀⠀⠀⠀⠛⠋⠀⠛⠀⠙⠋⠀⠛⠋⠈⠃⠈⠛⠁⠘⠛⠁", 0.8), lineDelay)
	printLine("", 0)

	time.Sleep(80 * time.Millisecond)
	printLine(colorizeRainbow("    ────────────────────────────────────────────────────", 0.5), lineDelay)
	time.Sleep(30 * time.Millisecond)
	printLine(colorizeRainbow(fmt.Sprintf("    DIRECCIÓN : %s", ip), 0.5), lineDelay)
	printLine(colorizeRainbow("    PUERTOS   : 6969 (TCP/WS)  ·  6970 (UDP)", 0.5), lineDelay)
	printLine(colorizeRainbow("    TOKEN SEG : "+activeToken, 0.5), lineDelay)
	time.Sleep(30 * time.Millisecond)
	printLine(colorizeRainbow("    ────────────────────────────────────────────────────", 0.5), lineDelay)
	time.Sleep(30 * time.Millisecond)
	printLine(colorizeRainbow("    [WiFi ]   Conecta la app AXON a la IP mostrada", 0.5), lineDelay)
	printLine(colorizeRainbow("    [USB  ]   Redireccionamiento ADB activo  →  puerto 6969", 0.5), lineDelay)
	printLine(colorizeRainbow("    [BT   ]   Usa la app AXON → modo Bluetooth desde Android", 0.5), lineDelay)
	time.Sleep(30 * time.Millisecond)
	printLine(colorizeRainbow("    ────────────────────────────────────────────────────", 0.5), lineDelay)
	printLine("", 0)

	time.Sleep(60 * time.Millisecond)
	fmt.Println(colorizeRainbow("    Escanea el QR con la app AXON:", 0.5))
	printLine("", 0)

	qrCode, err := qrcode.New(url, qrcode.Medium)
	if err == nil {
		printQRRainbow(qrCode.ToSmallString(false))
	} else {
		fmt.Println("Error al generar código QR en consola:", err)
	}

	fmt.Printf("\n")
	fmt.Println(colorizeRainbow(fmt.Sprintf("    O ingresa: %s", url), 0.5))
	fmt.Println()

	go startBluetoothServer()
	go startUDPServer()

	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		fmt.Println("Error al cargar archivos web estáticos:", err)
		os.Exit(1)
	}

	http.Handle("/", http.FileServer(http.FS(subFS)))
	http.HandleFunc("/ws", handleWebSocket)

	err = http.ListenAndServe(":6969", nil)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		os.Exit(1)
	}
}
