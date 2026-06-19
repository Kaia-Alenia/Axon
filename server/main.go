package main

import (
	"crypto/rand"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
)

//go:embed web/*
var webFS embed.FS

func setupADBReverse() {
	portStr := fmt.Sprintf("tcp:%d", port)
	if err := exec.Command("adb", "reverse", portStr, portStr).Run(); err != nil {
		fmt.Printf("[ADB] Warning: failed to setup adb reverse for port %d: %v\n", port, err)
	}
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
	activeClientIP    string
	activeClientNetIP net.IP
	activeClientMu    sync.RWMutex
	activeToken       string
	simulator         InputSimulator
	port              int
	udpPort           int
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
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
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

func isAddressLocal(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	if host == "127.0.0.1" || host == "::1" || host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() {
			return true
		}
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ipNet *net.IPNet
			switch v := addr.(type) {
			case *net.IPNet:
				ipNet = v
			case *net.IPAddr:
				ipNet = &net.IPNet{IP: v.IP, Mask: net.CIDRMask(len(v.IP)*8, len(v.IP)*8)}
			}
			if ipNet != nil && ipNet.IP.Equal(ip) {
				return true
			}
		}
	}
	return false
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	remoteAddrStr := r.RemoteAddr
	ip, _, _ := net.SplitHostPort(remoteAddrStr)
	isLocal := isAddressLocal(remoteAddrStr)

	if !isLocal && token != activeToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	if tcpConn, ok := conn.UnderlyingConn().(*net.TCPConn); ok {
		_ = tcpConn.SetNoDelay(true)
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(1 * time.Second)
		_ = tcpConn.SetReadBuffer(65536)
		_ = tcpConn.SetWriteBuffer(65536)
	}

	remoteAddrStr = conn.RemoteAddr().String()
	ip, _, _ = net.SplitHostPort(remoteAddrStr)

	activeClientMu.Lock()
	activeClientIP = ip
	activeClientNetIP = net.ParseIP(ip)
	activeClientMu.Unlock()

	fmt.Printf("[CONNECTION] Client connected from: %s\n", remoteAddrStr)

	done := make(chan bool)
	defer func() {
		close(done)
		activeClientMu.Lock()
		if activeClientIP == ip {
			activeClientIP = ""
			activeClientNetIP = nil
		}
		activeClientMu.Unlock()
		fmt.Println("[DISCONNECTION] Client disconnected")
	}()

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
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
					_ = conn.Close()
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
			fmt.Printf("[LATENCY] RTT Latency: %d ms\n", rtt)
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
			fmt.Printf("[KEYBOARD] Typing text: %q\n", msg.Text)
			simulator.Type(msg.Text)
		case "key":
			fmt.Printf("[KEYBOARD] Pressing special key: %q\n", msg.Key)
			simulator.Key(msg.Key)
		case "keycombo":
			fmt.Printf("[KEYBOARD] Pressing combo: %s+%s\n", msg.Modifier, msg.Key)
			simulator.KeyCombo(msg.Modifier, msg.Key)
		}
	}
}

func startUDPServer() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", udpPort))
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting UDP server:", err)
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
		clientNetIP := activeClientNetIP
		activeClientMu.RUnlock()

		if clientNetIP == nil || !remoteAddr.IP.Equal(clientNetIP) {
			continue
		}

		if n < 9 {
			continue
		}

		eventType := buf[0]

		xBits := uint32(buf[1])<<24 | uint32(buf[2])<<16 | uint32(buf[3])<<8 | uint32(buf[4])
		dx := math.Float32frombits(xBits)

		yBits := uint32(buf[5])<<24 | uint32(buf[6])<<16 | uint32(buf[7])<<8 | uint32(buf[8])
		dy := math.Float32frombits(yBits)

		switch eventType {
		case 0:
			simulator.MoveMouse(float64(dx), float64(dy))
		case 1:
			dir := "down"
			if dx > 0 {
				dir = "up"
			}
			simulator.Scroll(dir)
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

func startADBKeepalive() {
	portStr := fmt.Sprintf("tcp:%d", port)
	for {
		time.Sleep(3 * time.Second)
		activeClientMu.RLock()
		hasClient := activeClientIP != ""
		activeClientMu.RUnlock()
		if !hasClient {
			_ = exec.Command("adb", "reverse", portStr, portStr).Run()
		}
	}
}

func main() {
	flag.IntVar(&port, "port", 6969, "TCP/Websocket port to listen on")
	flag.IntVar(&udpPort, "udp-port", 6970, "UDP port to listen on")
	flag.Parse()

	go startADBKeepalive()
	simulator = NewSimulator()
	defer simulator.Close()

	setupADBReverse()

	ip := getLocalIP()
	if ip == "" {
		fmt.Println("Error: Could not retrieve local IP")
		os.Exit(1)
	}

	activeToken = generateToken()
	addr := fmt.Sprintf("%s:%d", ip, port)
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
	printLine(colorizeRainbow(fmt.Sprintf("    ADDRESS : %s", ip), 0.5), lineDelay)
	printLine(colorizeRainbow(fmt.Sprintf("    PORTS   : %d (TCP/WS)  ·  %d (UDP)", port, udpPort), 0.5), lineDelay)
	printLine(colorizeRainbow("    SEC TOKEN : "+activeToken, 0.5), lineDelay)
	time.Sleep(30 * time.Millisecond)
	printLine(colorizeRainbow("    ────────────────────────────────────────────────────", 0.5), lineDelay)
	time.Sleep(30 * time.Millisecond)
	printLine(colorizeRainbow("    [WiFi ]   Connect the AXON app to the displayed IP", 0.5), lineDelay)
	printLine(colorizeRainbow(fmt.Sprintf("    [USB  ]   Active ADB redirection  →  port %d", port), 0.5), lineDelay)
	printLine(colorizeRainbow("    [BT   ]   Use the AXON app → Bluetooth mode from Android", 0.5), lineDelay)
	time.Sleep(30 * time.Millisecond)
	printLine(colorizeRainbow("    ────────────────────────────────────────────────────", 0.5), lineDelay)
	printLine("", 0)

	time.Sleep(60 * time.Millisecond)
	fmt.Println(colorizeRainbow("    Scan the QR with the AXON app:", 0.5))
	printLine("", 0)

	qrCode, err := qrcode.New(url, qrcode.Medium)
	if err == nil {
		printQRRainbow(qrCode.ToSmallString(false))
	} else {
		fmt.Println("Error generating QR code in console:", err)
	}

	fmt.Printf("\n")
	fmt.Println(colorizeRainbow(fmt.Sprintf("    Or enter: %s", url), 0.5))
	fmt.Println()

	go startBluetoothServer()
	go startUDPServer()

	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		fmt.Println("Error loading static web files:", err)
		os.Exit(1)
	}

	http.Handle("/", http.FileServer(http.FS(subFS)))
	http.HandleFunc("/ws", handleWebSocket)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error starting the server:", err)
		os.Exit(1)
	}
}
