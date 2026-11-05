package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed web/*
var webFS embed.FS

func setupADBReverse() {
	if _, err := exec.LookPath("adb"); err != nil {
		return
	}
	portStr := fmt.Sprintf("tcp:%d", port)
	if err := exec.Command("adb", "reverse", portStr, portStr).Run(); err != nil {
		logger.Printf("[ADB] Warning: failed to setup adb reverse for port %d: exit status 1\n", port)
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
func isVirtualInterface(name string) bool {
	lower := strings.ToLower(name)
	virtualKeywords := []string{
		"vethernet", "wsl", "virtualbox", "vmware", "vbox",
		"docker", "virbr", "hyper-v", "teredo", "isatap",
		"ZeroTier", "tailscale", "hamachi", "npcap", "loopback",
	}
	for _, kw := range virtualKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}
func getLocalIP() string {
	if conn, err := net.DialTimeout("udp", "8.8.8.8:80", 300*time.Millisecond); err == nil {
		defer conn.Close()
		if udpAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			if ip4 := udpAddr.IP.To4(); ip4 != nil && !ip4.IsLoopback() && !ip4.IsLinkLocalUnicast() {
				return ip4.String()
			}
		}
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	priorityPrefixes := []string{"wi-fi", "wifi", "wireless", "wlan", "wlp", "wlo", "ethernet", "en", "eth"}
	for _, prefix := range priorityPrefixes {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			if isVirtualInterface(iface.Name) {
				continue
			}
			if !strings.HasPrefix(strings.ToLower(iface.Name), prefix) {
				continue
			}
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ip4 := ipnet.IP.To4(); ip4 != nil && !ip4.IsLoopback() && !ip4.IsLinkLocalUnicast() {
						return ip4.String()
					}
				}
			}
		}
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if isVirtualInterface(iface.Name) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ip4 := ipnet.IP.To4(); ip4 != nil && !ip4.IsLoopback() && !ip4.IsLinkLocalUnicast() {
					return ip4.String()
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
			if ip4 := ipnet.IP.To4(); ip4 != nil && !ip4.IsLinkLocalUnicast() {
				return ip4.String()
			}
		}
	}
	return ""
}
func createListenerWithReuseAddr(port int) (net.Listener, error) {
	lc := net.ListenConfig{
		Control: setReuseAddrControl,
	}
	return lc.Listen(context.Background(), "tcp", fmt.Sprintf(":%d", port))
}
func tryPortWithFallback(requestedPort int) (net.Listener, int, error) {
	listener, err := createListenerWithReuseAddr(requestedPort)
	if err == nil {
		return listener, requestedPort, nil
	}
	logger.Printf("[PORT] Port %d unavailable, attempting dynamic port allocation...\n", requestedPort)
	listener, err = createListenerWithReuseAddr(0)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to allocate any port: %v", err)
	}
	addr := listener.Addr().(*net.TCPAddr)
	logger.Printf("[PORT] Successfully allocated dynamic port: %d\n", addr.Port)
	return listener, addr.Port, nil
}

func tryUDPPortWithFallback(requestedPort int) (*net.UDPConn, int, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", requestedPort))
	if err == nil {
		conn, err := net.ListenUDP("udp", addr)
		if err == nil {
			return conn, requestedPort, nil
		}
	}
	logger.Printf("[PORT] UDP Port %d unavailable, attempting dynamic port allocation...\n", requestedPort)
	addr, err = net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to allocate any UDP port: %v", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to allocate any UDP port: %v", err)
	}
	actualPort := conn.LocalAddr().(*net.UDPAddr).Port
	logger.Printf("[PORT] Successfully allocated dynamic UDP port: %d\n", actualPort)
	return conn, actualPort, nil
}

var (
	activeClientIP    string
	activeClientNetIP net.IP
	activeClientMu    sync.RWMutex
	activeToken       string
	simulator         InputSimulator
	port              int
	udpPort           int
	logger            = newNonBlockingLogger()
)

type nonBlockingWriter struct {
	ch chan []byte
}

func newNonBlockingLogger() *log.Logger {
	w := &nonBlockingWriter{ch: make(chan []byte, 256)}
	go func() {
		for b := range w.ch {
			if isConsoleMinimized() {
				continue
			}
			_, _ = os.Stdout.Write(b)
		}
	}()
	return log.New(w, "", log.Ldate|log.Ltime)
}

func (w *nonBlockingWriter) Write(p []byte) (int, error) {
	buf := make([]byte, len(p))
	copy(buf, p)
	select {
	case w.ch <- buf:
	default:
	}
	return len(p), nil
}

type ClientMessage struct {
	Type      string  `json:"type"`
	Dx        float64 `json:"dx"`
	Dy        float64 `json:"dy"`
	Button    string  `json:"button"`
	Text      string  `json:"text"`
	Key       string  `json:"key"`
	Modifier  string  `json:"modifier"`
	Timestamp int64   `json:"timestamp"`
	Token     string  `json:"token,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		localIP := getLocalIP()
		allowed := []string{
			fmt.Sprintf("http://%s:%d", localIP, port),
			fmt.Sprintf("http://127.0.0.1:%d", port),
			fmt.Sprintf("http://localhost:%d", port),
			"http://" + r.Host,
		}
		for _, a := range allowed {
			if origin == a {
				return true
			}
		}
		if !strings.HasPrefix(origin, "http") {
			return true
		}
		logger.Printf("[SECURITY] Rejected WebSocket from origin: %s\n", origin)
		return false
	},
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
func handleAuthentication(w http.ResponseWriter, r *http.Request) bool {
	token := r.URL.Query().Get("token")
	if token != activeToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "" {
		if !handleAuthentication(w, r) {
			return
		}
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	if tcpConn, ok := conn.UnderlyingConn().(*net.TCPConn); ok {
		_ = tcpConn.SetNoDelay(true)
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(5 * time.Second)
		_ = tcpConn.SetReadBuffer(65536)
		_ = tcpConn.SetWriteBuffer(65536)
	}
	remoteAddrStr := conn.RemoteAddr().String()
	ip, _, _ := net.SplitHostPort(remoteAddrStr)
	activeClientMu.Lock()
	activeClientIP = ip
	activeClientNetIP = net.ParseIP(ip)
	activeClientMu.Unlock()
	logger.Printf("[CONNECTION] Client connected from: %s\n", remoteAddrStr)
	done := make(chan bool)
	defer func() {
		close(done)
		activeClientMu.Lock()
		if activeClientIP == ip {
			activeClientIP = ""
			activeClientNetIP = nil
		}
		activeClientMu.Unlock()
		logger.Println("[DISCONNECTION] Client disconnected")
	}()
	go startPingLoop(conn, done)
	handleMessages(conn)
}
func startPingLoop(conn *websocket.Conn, done chan bool) {
	ticker := time.NewTicker(5 * time.Second)
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
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err = conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				_ = conn.Close()
				return
			}
		case <-done:
			return
		}
	}
}
func handleMessages(conn *websocket.Conn) {
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
		if msg.Type == "pong" {
			continue
		}
		switch msg.Type {
		case "move":
			simulator.MoveMouse(msg.Dx, msg.Dy)
		case "click":
			logger.Printf("[MOUSE] Click: %s\n", msg.Button)
			simulator.Click(msg.Button)
		case "mousedown":
			logger.Printf("[MOUSE] MouseDown: %s\n", msg.Button)
			simulator.MouseDown(msg.Button)
		case "mouseup":
			logger.Printf("[MOUSE] MouseUp: %s\n", msg.Button)
			simulator.MouseUp(msg.Button)
		case "scroll":
			dir := "down"
			if msg.Dy > 0 {
				dir = "up"
			}
			simulator.Scroll(dir)
		case "type":
			logger.Printf("[KEYBOARD] Typing text: %q\n", msg.Text)
			simulator.Type(msg.Text)
		case "key":
			logger.Printf("[KEYBOARD] Pressing special key: %q\n", msg.Key)
			simulator.Key(msg.Key)
		case "keycombo":
			logger.Printf("[KEYBOARD] Pressing combo: %s+%s\n", msg.Modifier, msg.Key)
			simulator.KeyCombo(msg.Modifier, msg.Key)
		}
	}
}
func startUDPServer(conn *net.UDPConn) {
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
		if clientNetIP == nil || (!remoteAddr.IP.Equal(clientNetIP) && !remoteAddr.IP.IsPrivate() && !remoteAddr.IP.IsLoopback()) {
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
func startADBKeepalive() {
	if _, err := exec.LookPath("adb"); err != nil {
		return
	}
	portStr := fmt.Sprintf("tcp:%d", port)
	for {
		time.Sleep(3 * time.Second)
		activeClientMu.RLock()
		hasClient := activeClientIP != ""
		activeClientMu.RUnlock()
		if !hasClient {
			go func() {
				_ = exec.Command("adb", "reverse", portStr, portStr).Run()
			}()
		}
	}
}
func selfUpdate() {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	archName := goarch
	if goarch == "amd64" {
		archName = "amd64"
	} else if goarch == "arm64" {
		archName = "arm64"
	} else if goarch == "arm" {
		archName = "arm"
	}
	binaryName := fmt.Sprintf("axon-%s-%s", goos, archName)
	if goos == "windows" {
		binaryName += ".exe"
	}
	url := "https://github.com/Kaia-Alenia/Axon/releases/latest/download/" + binaryName
	logger.Printf("[UPDATE] Downloading latest version from:\n  %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		logger.Printf("[UPDATE] Error: failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Printf("[UPDATE] Error: server returned %s\n", resp.Status)
		os.Exit(1)
	}
	exePath, err := os.Executable()
	if err != nil {
		logger.Printf("[UPDATE] Error: could not determine executable path: %v\n", err)
		os.Exit(1)
	}
	tmpPath := exePath + ".new"
	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		logger.Printf("[UPDATE] Error: could not write update: %v\n", err)
		os.Exit(1)
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		logger.Printf("[UPDATE] Error: download failed: %v\n", err)
		_ = os.Remove(tmpPath)
		os.Exit(1)
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		logger.Printf("[UPDATE] Download complete. Could not auto-replace (file may be in use).\n")
		logger.Printf("[UPDATE] New binary saved at: %s\n", tmpPath)
		logger.Printf("[UPDATE] Manually replace the current binary with the downloaded file.\n")
		os.Exit(0)
	}
	logger.Println("[UPDATE] Update applied successfully. Restart axon to use the new version.")
	os.Exit(0)
}

func main() {
	showVersion := flag.Bool("version", false, "Show version information")
	runUpdate := flag.Bool("update", false, "Download and install the latest version of Axon")
	noLogo := flag.Bool("no-logo", false, "Skip the decorative logo (use if your terminal font lacks braille support)")
	minimized := flag.Bool("minimized", false, "Start Axon with the console window minimized")
	flag.IntVar(&port, "port", 6969, "TCP/WebSocket port to listen on (0 for auto-allocation)")
	flag.IntVar(&udpPort, "udp-port", 6970, "UDP port for fast input protocol (0 for auto-allocation)")
	flag.Parse()
	if *showVersion {
		fmt.Println("Axon Server v1.6.9")
		fmt.Println("High-speed wireless input bridge")
		os.Exit(0)
	}
	if *runUpdate {
		selfUpdate()
	}
	if *minimized {
		minimizeConsoleWindow()
	}
	go startADBKeepalive()
	simulator = NewSimulator()
	defer simulator.Close()
	setupADBReverse()
	ip := getLocalIP()
	if ip == "" {
		logger.Println("Error: Could not retrieve local IP address")
		os.Exit(1)
	}
	tcpListener, actualPort, err := tryPortWithFallback(port)
	if err != nil {
		logger.Printf(" Error: Failed to bind to any port: %v\n", err)
		os.Exit(1)
	}
	defer tcpListener.Close()
	if actualPort != port {
		logger.Printf("[PORT] Using dynamic port: %d (requested: %d)\n", actualPort, port)
		port = actualPort
	}

	udpConn, actualUdpPort, err := tryUDPPortWithFallback(udpPort)
	if err != nil {
		logger.Printf(" Error: Failed to bind UDP port: %v\n", err)
		os.Exit(1)
	}
	if actualUdpPort != udpPort {
		logger.Printf("[PORT] Using dynamic UDP port: %d (requested: %d)\n", actualUdpPort, udpPort)
		udpPort = actualUdpPort
	}
	activeToken = generateToken()
	addr := fmt.Sprintf("%s:%d", ip, port)
	url := fmt.Sprintf("http://%s/?token=%s&udp=%d", addr, activeToken, udpPort)
	lineDelay := 55 * time.Millisecond
	printLine(colorizeRainbow("    ___  _  ______  _  __", 0.4), lineDelay)
	printLine(colorizeRainbow("   /   || |/ / __ \\/ |/ /", 0.4), lineDelay)
	printLine(colorizeRainbow("  / /| ||   / / / /    / ", 0.4), lineDelay)
	printLine(colorizeRainbow(" / ___ |/   / /_/ /    /  ", 0.4), lineDelay)
	printLine(colorizeRainbow("/_/  |_/_/|_\\____/_/|_/   ", 0.4), lineDelay)
	printLine(colorizeRainbow("    [ HIGH-SPEED INPUT BRIDGE ]", 0.6), lineDelay)
	if !*noLogo {
		printLine(colorizeRainbow("               ___    __    ____ _  __ ____ ___ ", 0.8), lineDelay)
		printLine(colorizeRainbow("              /   |  / /   / __/ / / //  _//   |", 0.8), lineDelay)
		printLine(colorizeRainbow("             / /| | / /   / /_  /  / // / / /| |", 0.8), lineDelay)
		printLine(colorizeRainbow("            / ___ |/ /___/ __/ / /  // / / ___ |", 0.8), lineDelay)
		printLine(colorizeRainbow("           /_/  |_/_____/___/ /_/ /_/___/_/  |_|", 0.8), lineDelay)
		printLine(colorizeRainbow("                   S T U D I O S", 0.8), lineDelay)
	}
	printLine("", 0)

	printLine(colorizeRainbow("    ────────────────────────────────────────────────────", 0.5), lineDelay)

	time.Sleep(30 * time.Millisecond)
	printLine(colorizeRainbow(fmt.Sprintf("     ADDRESS : %s", ip), 0.5), lineDelay)
	printLine(colorizeRainbow(fmt.Sprintf("     PORTS   : %d (TCP/WS)  ·  %d (UDP)", port, udpPort), 0.5), lineDelay)
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
	fmt.Println(colorizeRainbow("     Scan the QR with the AXON app:", 0.5))
	printLine("", 0)
	printTerminalQRCode(url)
	fmt.Printf("\n")
	fmt.Println(colorizeRainbow(fmt.Sprintf("    Or enter: %s", url), 0.5))
	fmt.Println()
	go startBluetoothServer()
	go startUDPServer(udpConn)
	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		logger.Println(" Error loading static web files:", err)
		os.Exit(1)
	}
	http.Handle("/", http.FileServer(http.FS(subFS)))
	http.HandleFunc("/ws", handleWebSocket)
	server := &http.Server{}
	logger.Printf("[SERVER]  Listening on port %d...\n", port)
	err = server.Serve(tcpListener)
	if err != nil {
		logger.Printf(" Error starting the server: %v\n", err)
		os.Exit(1)
	}
}
