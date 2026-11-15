package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/json"
	"errors"
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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed web/*
var webFS embed.FS

// version is the server version. Keep in sync with the Android
// versionName (1.7.3, versionCode 6).
const version = "1.7.3"

const githubLatestReleaseAPI = "https://api.github.com/repos/Kaia-Alenia/Axon/releases/latest"

// updateCheckTimeout bounds the GitHub API version check so a dead
// network fails fast instead of hanging the update.
const updateCheckTimeout = 10 * time.Second

// updateDownloadTimeout bounds the binary download.
const updateDownloadTimeout = 120 * time.Second

// updateDiskSpaceMargin is extra free space required on top of the new
// binary itself, for the .new file, the .old backup and filesystem
// overhead, so a download never fills the disk exactly to zero.
const updateDiskSpaceMargin = 1 << 20 // 1 MiB

// Win32 error codes used to tell a transient file lock apart from a
// real permission problem. On Windows an antivirus or another process
// holding the exe very often surfaces as ERROR_ACCESS_DENIED (5),
// not as a sharing violation, so 5 is treated as a lock (retryable)
// here; a genuine permission problem is caught earlier by the
// install-directory writability probe (see checkInstallDirWritable).
const (
	windowsERROR_ACCESS_DENIED     = 5  // ERROR_ACCESS_DENIED
	windowsERROR_SHARING_VIOLATION = 32 // ERROR_SHARING_VIOLATION
	windowsERROR_LOCK_VIOLATION    = 33 // ERROR_LOCK_VIOLATION
)

// Downloaded binaries smaller than this are not plausible Axon builds;
// larger than maxDownloadSize they are not plausible either.
const (
	minDownloadSize = 512
	maxDownloadSize = 200 << 20 // 200 MiB
)

// updateSwapTimeout / updateSwapInterval control how long the update
// helper polls for the running binary to become replaceable. They are
// variables (not consts) so tests can shorten the wait.
var (
	updateSwapTimeout  = 10 * time.Second
	updateSwapInterval = 200 * time.Millisecond
	// updateSwapFallbackTimeout bounds the retry loop for the
	// rename-away fallback inside waitAndSwap (see below).
	updateSwapFallbackTimeout = 20 * time.Second
	// updateSwapMaxInterval caps the backoff between fallback attempts.
	updateSwapMaxInterval = 2 * time.Second
	// updateLaunchCheckTimeout bounds the "--version" probe that proves
	// a freshly swapped binary actually launches.
	updateLaunchCheckTimeout = 10 * time.Second
)

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
	ch       chan []byte
	flushReq chan chan struct{}
}

func newNonBlockingLogger() *log.Logger {
	w := &nonBlockingWriter{ch: make(chan []byte, 256), flushReq: make(chan chan struct{})}
	go func() {
		for {
			select {
			case b := <-w.ch:
				if isConsoleMinimized() {
					continue
				}
				_, _ = os.Stdout.Write(b)
			case done := <-w.flushReq:
				// Drain everything queued behind the flush request, in
				// order, then signal completion.
			drain:
				for {
					select {
					case b := <-w.ch:
						if isConsoleMinimized() {
							continue
						}
						_, _ = os.Stdout.Write(b)
					default:
						close(done)
						break drain
					}
				}
			}
		}
	}()
	return log.New(w, "", log.Ldate|log.Ltime)
}

// flushLogger blocks until every log line queued so far has been
// written to stdout. Call it before os.Exit: the background writer
// goroutine does not run during process exit, so without a flush the
// last lines (e.g. "Already up to date") are silently lost. This bit
// the self-updater, which exits right after logging its result —
// "axon --update" on an up-to-date install printed nothing at all.
func flushLogger() {
	if w, ok := logger.Writer().(*nonBlockingWriter); ok {
		done := make(chan struct{})
		w.flushReq <- done
		<-done
	}
}

// exitUpdate flushes pending log output and then exits with code.
// Use it instead of os.Exit on every path out of the self-update
// flow, so the user actually sees the result.
func exitUpdate(code int) {
	flushLogger()
	os.Exit(code)
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
// normalizeVersionTag strips surrounding whitespace and a leading "v"
// ("v1.6.9" -> "1.6.9") so release tags compare against version.
func normalizeVersionTag(tag string) string {
	tag = strings.TrimSpace(tag)
	tag = strings.TrimPrefix(tag, "v")
	tag = strings.TrimPrefix(tag, "V")
	return tag
}

// parseSemver parses "1.2.3" (1-3 numeric parts; a leading "v" and any
// "-prerelease"/"+build" suffix are ignored) into [major minor patch].
func parseSemver(s string) ([3]int, error) {
	var out [3]int
	s = normalizeVersionTag(s)
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	if s == "" {
		return out, fmt.Errorf("empty version")
	}
	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return out, fmt.Errorf("invalid version %q: too many parts", s)
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, fmt.Errorf("invalid version %q: part %q is not a non-negative number", s, p)
		}
		out[i] = n
	}
	return out, nil
}

// compareVersions compares two semver strings numerically: -1 if a < b,
// 0 if equal, +1 if a > b. It returns an error if either side is not a
// valid version.
func compareVersions(a, b string) (int, error) {
	av, err := parseSemver(a)
	if err != nil {
		return 0, err
	}
	bv, err := parseSemver(b)
	if err != nil {
		return 0, err
	}
	for i := 0; i < 3; i++ {
		if av[i] < bv[i] {
			return -1, nil
		}
		if av[i] > bv[i] {
			return 1, nil
		}
	}
	return 0, nil
}

type githubReleaseAsset struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type githubRelease struct {
	TagName string               `json:"tag_name"`
	Assets  []githubReleaseAsset `json:"assets"`
}

// fetchLatestReleaseInfo asks the GitHub API for the latest release and
// returns its tag as a plain version ("v1.7.0" -> "1.7.0") plus the size
// in bytes of the asset named binaryName, or -1 when the API lists no
// such asset (unknown size, not an error). apiURL is a parameter
// (instead of the const) so tests can point it at a local test server.
func fetchLatestReleaseInfo(client *http.Client, apiURL, binaryName string) (string, int64, error) {
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", -1, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", -1, fmt.Errorf("GitHub API returned %s", resp.Status)
	}
	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", -1, fmt.Errorf("could not parse GitHub API response: %v", err)
	}
	if strings.TrimSpace(rel.TagName) == "" {
		return "", -1, fmt.Errorf("GitHub API response has no tag_name")
	}
	assetSize := int64(-1)
	for _, a := range rel.Assets {
		if binaryName != "" && a.Name == binaryName {
			assetSize = a.Size
			break
		}
	}
	return normalizeVersionTag(rel.TagName), assetSize, nil
}

// fetchLatestReleaseVersion asks the GitHub API for the latest release
// tag and returns it as a plain version ("v1.7.0" -> "1.7.0"). apiURL is
// a parameter (instead of the const) so tests can point it at a local
// test server.
func fetchLatestReleaseVersion(client *http.Client, apiURL string) (string, error) {
	tag, _, err := fetchLatestReleaseInfo(client, apiURL, "")
	return tag, err
}

// validateDownloadedBinary checks a downloaded file is a plausible
// executable for goos: non-empty, sane size, and the right magic bytes
// (MZ on Windows, ELF on Linux, Mach-O on macOS). Unknown platforms get
// the size check only.
func validateDownloadedBinary(path, goos string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("could not stat downloaded file: %v", err)
	}
	if info.Size() < minDownloadSize {
		return fmt.Errorf("downloaded file is suspiciously small (%d bytes)", info.Size())
	}
	if info.Size() > maxDownloadSize {
		return fmt.Errorf("downloaded file is suspiciously large (%d bytes)", info.Size())
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open downloaded file: %v", err)
	}
	defer f.Close()
	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil {
		return fmt.Errorf("could not read downloaded file: %v", err)
	}
	switch goos {
	case "windows":
		if magic[0] != 'M' || magic[1] != 'Z' {
			return fmt.Errorf("downloaded file is not a Windows executable (missing MZ header)")
		}
	case "linux":
		if magic[0] != 0x7f || magic[1] != 'E' || magic[2] != 'L' || magic[3] != 'F' {
			return fmt.Errorf("downloaded file is not a Linux executable (missing ELF header)")
		}
	case "darwin":
		machO := (magic[0] == 0xfe && magic[1] == 0xed && magic[2] == 0xfa && (magic[3] == 0xce || magic[3] == 0xcf)) ||
			(magic[0] == 0xce && magic[1] == 0xfa && magic[2] == 0xed && magic[3] == 0xfe) ||
			(magic[0] == 0xcf && magic[1] == 0xfa && magic[2] == 0xed && magic[3] == 0xfe) ||
			(magic[0] == 0xca && magic[1] == 0xfe && magic[2] == 0xba && (magic[3] == 0xbe || magic[3] == 0xbf))
		if !machO {
			return fmt.Errorf("downloaded file is not a macOS executable (missing Mach-O header)")
		}
	}
	return nil
}

// formatBytes renders a byte count for user-facing messages.
func formatBytes(n uint64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d B", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f KiB", float64(n)/1024)
	case n < 1024*1024*1024:
		return fmt.Sprintf("%.1f MiB", float64(n)/(1024*1024))
	default:
		return fmt.Sprintf("%.2f GiB", float64(n)/(1024*1024*1024))
	}
}

// errnoOf unwraps err to a syscall.Errno when one is in the chain.
func errnoOf(err error) (syscall.Errno, bool) {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno, true
	}
	return 0, false
}

// isPermissionDeniedError reports whether err is a permission problem
// (os.ErrPermission covers the platform mapping; on Windows the raw
// ERROR_ACCESS_DENIED code is checked too).
func isPermissionDeniedError(err error) bool {
	if errors.Is(err, os.ErrPermission) {
		return true
	}
	if errno, ok := errnoOf(err); ok && runtime.GOOS == "windows" && errno == windowsERROR_ACCESS_DENIED {
		return true
	}
	return false
}

// isFileLockedError reports whether err looks like a transient file
// lock (another process or antivirus holding the binary). Only
// meaningful on Windows: on Unix a running binary is not locked
// against rename, so there a denied rename is a permission problem.
// Note ERROR_ACCESS_DENIED is included because antivirus locks very
// often surface as access-denied rather than as a sharing violation.
func isFileLockedError(err error) bool {
	if runtime.GOOS != "windows" {
		return false
	}
	if errno, ok := errnoOf(err); ok {
		switch errno {
		case windowsERROR_SHARING_VIOLATION, windowsERROR_LOCK_VIOLATION, windowsERROR_ACCESS_DENIED:
			return true
		}
	}
	return false
}

// permissionHint tells the user the one action that fixes a genuine
// permission problem.
func permissionHint() string {
	if runtime.GOOS == "windows" {
		return "try running Axon as Administrator (right-click the exe > Run as administrator) or move Axon to a folder you can write to"
	}
	return "check the directory permissions or move Axon to a folder you can write to"
}

// checkFreeDiskSpace fails when dir does not have requiredBytes free.
// A failure to even measure the free space only logs a warning and
// lets the update continue: the download is still verified
// byte-for-byte afterwards, so an exotic filesystem must not block
// updates outright.
func checkFreeDiskSpace(dir string, requiredBytes int64) error {
	free, err := freeDiskSpace(dir)
	if err != nil {
		logger.Printf("[UPDATE] Warning: could not check free disk space in %s: %v\n", dir, err)
		return nil
	}
	if requiredBytes > 0 && uint64(requiredBytes) > free {
		return fmt.Errorf("not enough free disk space in %s: need at least %s free, only %s available — free some space and try again", dir, formatBytes(uint64(requiredBytes)), formatBytes(free))
	}
	return nil
}

// requiredDownloadBytes is how much free space the download needs:
// the expected binary size when known (GitHub asset size), otherwise
// the current exe size as the best estimate, plus a margin in both
// cases for the .new file, the .old backup and filesystem overhead.
func requiredDownloadBytes(expectedSize int64, exePath string) int64 {
	if expectedSize > 0 {
		return expectedSize + updateDiskSpaceMargin
	}
	if info, err := os.Stat(exePath); err == nil {
		return info.Size() + updateDiskSpaceMargin
	}
	return updateDiskSpaceMargin
}

// checkInstallDirWritable probes the install directory with a real
// create+delete before anything is downloaded, so a read-only
// directory or a permission problem fails fast with a clear message
// instead of after a wasted download. Mode bits alone do not prove
// writability (ACLs, read-only mounts, Windows attributes), which is
// why this actually writes a probe file.
func checkInstallDirWritable(dir string) error {
	if info, err := os.Stat(dir); err != nil {
		return fmt.Errorf("install directory %s is not accessible: %w", dir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("install directory %s is not a directory", dir)
	}
	f, err := os.CreateTemp(dir, ".axon-update-probe-*")
	if err != nil {
		hint := ""
		if isPermissionDeniedError(err) {
			hint = " — the install directory is read-only or permission was denied; " + permissionHint()
		}
		return fmt.Errorf("install directory %s is not writable: %w%s", dir, err, hint)
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return nil
}

// verifyDownloadedSize checks a downloaded file is exactly
// expectedSize bytes. expectedSize <= 0 means unknown: no check.
func verifyDownloadedSize(path string, expectedSize int64) error {
	if expectedSize <= 0 {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("could not stat downloaded file: %w", err)
	}
	if info.Size() != expectedSize {
		return fmt.Errorf("downloaded file is %d bytes, expected %d bytes (incomplete or corrupt download)", info.Size(), expectedSize)
	}
	return nil
}

// parseVersionFromOutput finds the first plausible "X.Y.Z" version in
// program output such as "Axon Server v1.7.0". A bare number with no
// dot (a year, a port) does not count.
func parseVersionFromOutput(output string) (string, error) {
	for _, field := range strings.Fields(output) {
		candidate := strings.Trim(field, ",;:()[]\"'")
		if !strings.Contains(candidate, ".") {
			continue
		}
		if _, err := parseSemver(candidate); err == nil {
			return normalizeVersionTag(candidate), nil
		}
	}
	return "", fmt.Errorf("no version number found in output %q", strings.TrimSpace(output))
}

// verifyExeVersion proves a binary actually launches: it runs
// "<exe> --version" and checks the printed version. When
// expectedVersion is non-empty the printed version must equal it;
// when it is empty (version check unavailable) any parseable version
// passes — the point is that the binary starts and reports itself.
func verifyExeVersion(exePath, expectedVersion string) error {
	ctx, cancel := context.WithTimeout(context.Background(), updateLaunchCheckTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, exePath, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("launch check timed out after %s: %s did not respond to --version", updateLaunchCheckTimeout, exePath)
		}
		return fmt.Errorf("launch check failed: %s --version exited with %v (output: %q)", exePath, err, strings.TrimSpace(string(out)))
	}
	got, err := parseVersionFromOutput(string(out))
	if err != nil {
		return fmt.Errorf("launch check failed: %s started but %v (output: %q)", exePath, err, strings.TrimSpace(string(out)))
	}
	if expectedVersion != "" {
		cmp, cmpErr := compareVersions(got, expectedVersion)
		if cmpErr != nil {
			return fmt.Errorf("launch check failed: could not compare printed version %q with expected %q: %v", got, expectedVersion, cmpErr)
		}
		if cmp != 0 {
			return fmt.Errorf("launch check failed: new binary reports version %s, expected %s", got, expectedVersion)
		}
	}
	return nil
}

// backupCurrentExe copies the running binary to <exe>.old. It is the
// backup for the direct-rename path, which would otherwise destroy
// the old binary with no way back; the rename-away swap creates the
// same <exe>.old file by moving instead of copying, so there is one
// backup file and one scheme, not two.
func backupCurrentExe(exePath string) error {
	oldPath := exePath + ".old"
	src, err := os.Open(exePath)
	if err != nil {
		return fmt.Errorf("could not open running binary for backup: %w", err)
	}
	defer src.Close()
	info, err := src.Stat()
	if err != nil {
		return fmt.Errorf("could not stat running binary for backup: %w", err)
	}
	perm := info.Mode().Perm()
	if perm == 0 {
		perm = 0755
	}
	dst, err := os.OpenFile(oldPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("could not create backup %s: %w", oldPath, err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		_ = os.Remove(oldPath)
		return fmt.Errorf("could not copy running binary to backup %s: %w", oldPath, err)
	}
	if err := dst.Sync(); err != nil {
		_ = dst.Close()
		_ = os.Remove(oldPath)
		return fmt.Errorf("could not finish writing backup %s: %w", oldPath, err)
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(oldPath)
		return fmt.Errorf("could not finish writing backup %s: %w", oldPath, err)
	}
	return nil
}

// restoreFromBackup puts <exe>.old back over exePath after a failed
// swap or a failed launch check. On success the backup is consumed
// (it has been renamed back). On failure it reports both the failure
// and the exact manual step, and leaves the backup in place.
func restoreFromBackup(exePath string) error {
	oldPath := exePath + ".old"
	if _, err := os.Stat(oldPath); err != nil {
		return fmt.Errorf("no backup to restore from: %s is missing (%v); the old binary cannot be recovered automatically — reinstall Axon from the latest release", oldPath, err)
	}
	// os.Rename replaces the destination on both Unix and Windows, so
	// this works even if the broken new binary is still at exePath.
	if renameErr := os.Rename(oldPath, exePath); renameErr == nil {
		return nil
	} else if removeErr := os.Remove(exePath); removeErr != nil {
		return fmt.Errorf("could not restore backup: renaming %s over %s failed: %v, and removing the broken binary also failed: %v — restore it manually by copying %s over %s", oldPath, exePath, renameErr, removeErr, oldPath, exePath)
	} else if renameErr2 := os.Rename(oldPath, exePath); renameErr2 != nil {
		return fmt.Errorf("could not restore backup: renaming %s over %s failed after removing the broken binary: %v (first attempt: %v) — the backup is still at %s; copy it over %s manually", oldPath, exePath, renameErr2, renameErr, oldPath, exePath)
	}
	return nil
}

// cleanupStaleBackup removes a leftover <exe>.old on normal startup,
// but ONLY after the current binary has proven it launches (the same
// --version probe the updater uses). If that probe fails, the current
// binary may be a broken update, so the backup is KEPT: it is the
// only way back to a working Axon. A previous version deleted .old
// unconditionally, which could destroy the rollback copy of an
// update that had swapped but never verified.
func cleanupStaleBackup(exePath string) {
	oldPath := exePath + ".old"
	if _, err := os.Stat(oldPath); err != nil {
		return // No backup left behind; nothing to clean up.
	}
	if err := verifyExeVersion(exePath, ""); err != nil {
		logger.Printf("[UPDATE] Keeping backup %s: the current binary did not pass the launch check (%v). If Axon misbehaves, restore it manually by copying %s over %s.\n", oldPath, err, oldPath, exePath)
		return
	}
	if err := os.Remove(oldPath); err != nil {
		logger.Printf("[UPDATE] Warning: could not remove stale backup %s: %v\n", oldPath, err)
		return
	}
	logger.Printf("[UPDATE] Removed stale backup %s (current binary verified working).\n", oldPath)
}

// appendUpdateLog appends one line to update.log next to the exe so a
// failed update can be diagnosed after the fact.
func appendUpdateLog(exePath, msg string) {
	logPath := filepath.Join(filepath.Dir(exePath), "update.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.Printf("[UPDATE] Warning: could not write %s: %v\n", logPath, err)
		return
	}
	_, _ = fmt.Fprintf(f, "[%s] %s\n", time.Now().Format(time.RFC3339), msg)
	_ = f.Close()
}

// downloadUpdateExpected fetches url into tmpPath and verifies the
// transfer: bytes written must equal the HTTP Content-Length when the
// server sends one, and expectedSize (from the releases API) when
// known (expectedSize <= 0 means unknown). A mismatch means the
// download was truncated or corrupt: the partial file is deleted and
// an error is returned, never a half-written binary. It returns the
// number of bytes written. Compression is disabled so Content-Length
// (bytes on the wire) matches the bytes written to disk.
func downloadUpdateExpected(url, tmpPath string, expectedSize int64) (int64, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableCompression = true
	client := &http.Client{
		Timeout:   updateDownloadTimeout,
		Transport: transport,
	}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to connect: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("server returned %s", resp.Status)
	}
	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return 0, fmt.Errorf("could not write update: %v", err)
	}
	written, copyErr := io.Copy(out, resp.Body)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return written, fmt.Errorf("download failed after %d bytes: %v", written, copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return written, fmt.Errorf("could not finish writing update: %v", closeErr)
	}
	if resp.ContentLength >= 0 && written != resp.ContentLength {
		_ = os.Remove(tmpPath)
		return written, fmt.Errorf("incomplete download: received %d bytes, server said %d (Content-Length); partial file deleted", written, resp.ContentLength)
	}
	if expectedSize > 0 && written != expectedSize {
		_ = os.Remove(tmpPath)
		return written, fmt.Errorf("incomplete download: received %d bytes, expected %d bytes from the release; partial file deleted", written, expectedSize)
	}
	return written, nil
}

// downloadUpdate fetches url into tmpPath. It verifies the transfer
// against Content-Length (see downloadUpdateExpected); callers must
// still run validateDownloadedBinary afterwards.
func downloadUpdate(url, tmpPath string) error {
	_, err := downloadUpdateExpected(url, tmpPath, -1)
	return err
}

// launchNewExe starts exePath as a new process (plain, no flags) with
// the terminal attached, and returns once it is running. Used by the
// update helper to relaunch Axon after the swap.
func launchNewExe(exePath string) error {
	cmd := exec.Command(exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start %s: %v", exePath, err)
	}
	return nil
}

// writeUpdateLog records a failed swap next to the executable so the
// user can see what happened without re-running the update. It
// appends, so the steps that led to the failure stay in the file.
func writeUpdateLog(exePath string, updateErr error) {
	appendUpdateLog(exePath, fmt.Sprintf("Update failed: %v", updateErr))
	logger.Printf("[UPDATE] Details written to: %s\n", filepath.Join(filepath.Dir(exePath), "update.log"))
}

// swapViaRenameAway is the Windows fallback for replacing a running
// executable: a running .exe cannot be overwritten, but it CAN be
// renamed, so the old binary is moved to <exe>.old and the new binary
// takes its place. It is also the backup-first swap used on every
// platform (see selfUpdate): the old binary is never destroyed
// without a copy at <exe>.old. The .old file is removed only after
// the new binary has passed the launch check (see applyUpdate and
// cleanupStaleBackup); on failure it tries to restore the old binary.
func swapViaRenameAway(exePath, newPath string) error {
	oldPath := exePath + ".old"
	_ = os.Remove(oldPath)
	if err := os.Rename(exePath, oldPath); err != nil {
		return fmt.Errorf("could not move running binary aside: %w", err)
	}
	if err := os.Rename(newPath, exePath); err != nil {
		_ = os.Rename(oldPath, exePath)
		return fmt.Errorf("could not move new binary into place: %w", err)
	}
	return nil
}

// waitAndSwap replaces exePath with newPath, polling the rename until
// the process that holds the binary has exited (updateSwapTimeout,
// every updateSwapInterval). If the direct rename keeps failing — the
// Windows case, where the running exe is locked — it falls back to
// swapViaRenameAway, retried with backoff until
// updateSwapFallbackTimeout: a lock held by antivirus or another
// process usually clears within seconds, so one attempt is not
// enough. A genuine permission error is not retried — waiting will
// not fix it — and is reported with the Administrator hint.
func waitAndSwap(parentPID int, exePath, newPath string) error {
	logger.Printf("[UPDATE] Waiting for process %d to exit before replacing the binary...\n", parentPID)
	deadline := time.Now().Add(updateSwapTimeout)
	var lastErr error
	for {
		if err := os.Rename(newPath, exePath); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(updateSwapInterval)
	}
	logger.Printf("[UPDATE] Direct replace still failing (%v); trying to move the running binary aside...\n", lastErr)
	fallbackDeadline := time.Now().Add(updateSwapFallbackTimeout)
	interval := updateSwapInterval
	var fallbackErr error
	for {
		if err := swapViaRenameAway(exePath, newPath); err == nil {
			return nil
		} else {
			fallbackErr = err
		}
		if isPermissionDeniedError(fallbackErr) && !isFileLockedError(fallbackErr) {
			return fmt.Errorf("could not replace %s after %s: %v; fallback also failed: %w — permission denied; %s", exePath, updateSwapTimeout, lastErr, fallbackErr, permissionHint())
		}
		if time.Now().After(fallbackDeadline) {
			break
		}
		logger.Printf("[UPDATE] Swap still failing (%v); retrying in %s (file may be locked by another process or antivirus)...\n", fallbackErr, interval)
		time.Sleep(interval)
		interval *= 2
		if interval > updateSwapMaxInterval {
			interval = updateSwapMaxInterval
		}
	}
	return fmt.Errorf("could not replace %s after %s: %v; fallback also failed: %w — the file may still be locked by another process or antivirus, or permission may be denied; close other programs using Axon and try again, or %s", exePath, updateSwapTimeout, lastErr, fallbackErr, permissionHint())
}

// applyUpdate is the hidden helper run as
// "axon --apply-update <parent-pid> <exe-path> <new-path> [expected-version]".
// It waits for the parent to exit, swaps the new binary over the old
// one, verifies the new binary actually launches, and relaunches Axon.
// It never runs the server. The <exe-path> and <new-path> arguments
// arrive as separate argv entries (see selfUpdate), so paths with
// spaces need no quoting and must never be joined into one command
// string. On failure it writes update.log next to the exe, restores
// the old binary from the backup when possible, and leaves <exe>.new
// in place only when the swap itself never happened.
func applyUpdate(parentPID int, exePath, newPath string) error {
	return applyUpdateExpected(parentPID, exePath, newPath, "")
}

func applyUpdateExpected(parentPID int, exePath, newPath, expectedVersion string) error {
	if parentPID <= 0 {
		return fmt.Errorf("invalid parent pid %d", parentPID)
	}
	// Backup first: the direct rename inside waitAndSwap would
	// otherwise destroy the old binary with no way back. This copy
	// and the rename-away fallback use the same <exe>.old file (one
	// backup, one scheme). If the copy fails the swap may still
	// create .old by renaming, so this is a warning, not a fatal
	// error — but it is logged, because a failed launch check with
	// no backup cannot be rolled back.
	if err := backupCurrentExe(exePath); err != nil {
		appendUpdateLog(exePath, fmt.Sprintf("Warning: could not back up the current binary before swapping: %v", err))
		logger.Printf("[UPDATE] Warning: could not back up the current binary: %v\n", err)
	}
	if err := waitAndSwap(parentPID, exePath, newPath); err != nil {
		writeUpdateLog(exePath, err)
		return err
	}
	appendUpdateLog(exePath, "Swap complete; verifying the new binary launches...")
	if err := verifyExeVersion(exePath, expectedVersion); err != nil {
		// The new binary is in place but does not run: roll back to
		// the backup instead of leaving the user with a broken Axon.
		restoreErr := restoreFromBackup(exePath)
		if restoreErr != nil {
			combined := fmt.Errorf("%v; rollback also failed: %v", err, restoreErr)
			writeUpdateLog(exePath, combined)
			return combined
		}
		restored := fmt.Errorf("%v; old version restored from backup", err)
		writeUpdateLog(exePath, restored)
		return restored
	}
	appendUpdateLog(exePath, "Launch check passed; restarting axon...")
	logger.Println("[UPDATE] Update applied successfully. Restarting axon...")
	if err := launchNewExe(exePath); err != nil {
		launchErr := fmt.Errorf("update applied but could not restart axon: %v (start %s manually)", err, exePath)
		writeUpdateLog(exePath, launchErr)
		return launchErr
	}
	// Only now, with the new binary verified AND relaunched, is the
	// backup disposable. If this remove fails, the backup simply
	// lingers; the next normal startup removes it after its own
	// launch check (see cleanupStaleBackup).
	if err := os.Remove(exePath + ".old"); err != nil && !os.IsNotExist(err) {
		appendUpdateLog(exePath, fmt.Sprintf("Warning: could not remove backup %s: %v", exePath+".old", err))
		logger.Printf("[UPDATE] Warning: could not remove backup %s: %v\n", exePath+".old", err)
	}
	return nil
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
	exePath, err := os.Executable()
	if err != nil {
		logger.Printf("[UPDATE] Error: could not determine executable path: %v\n", err)
		exitUpdate(1)
	}
	exeDir := filepath.Dir(exePath)
	tmpPath := exePath + ".new"
	// Remove any stale download from a previous failed update.
	_ = os.Remove(tmpPath)

	// Version check first, before any writability probe: an
	// up-to-date install must report "Already up to date" even when
	// it lives in a read-only directory (e.g. /usr/local/bin), where
	// the probe below would otherwise fail first and hide the fact
	// that no update was needed at all.
	apiClient := &http.Client{Timeout: updateCheckTimeout}
	latest := ""
	expectedSize := int64(-1)
	latest, expectedSize, err = fetchLatestReleaseInfo(apiClient, githubLatestReleaseAPI, binaryName)
	if err != nil {
		logger.Printf("[UPDATE] Warning: could not check latest version (%v); downloading latest release anyway.\n", err)
		latest = ""
		expectedSize = -1
	} else if cmp, cmpErr := compareVersions(latest, version); cmpErr != nil {
		logger.Printf("[UPDATE] Warning: could not compare versions (latest %q): %v; downloading latest release anyway.\n", latest, cmpErr)
	} else if cmp == 0 {
		logger.Printf("[UPDATE] Already up to date (v%s).\n", version)
		exitUpdate(0)
	} else if cmp < 0 {
		logger.Printf("[UPDATE] You are running v%s, newer than the latest release (v%s); nothing to do.\n", version, latest)
		exitUpdate(0)
	} else {
		logger.Printf("[UPDATE] New version available: v%s (current v%s).\n", latest, version)
	}

	// Fail fast when the install directory cannot be written (read-only
	// directory, permission denied) instead of after a wasted download.
	if err := checkInstallDirWritable(exeDir); err != nil {
		logger.Printf("[UPDATE] Error: %v\n", err)
		appendUpdateLog(exePath, fmt.Sprintf("Update failed: %v", err))
		exitUpdate(1)
	}

	// Refuse to start a download that cannot fit, instead of producing
	// a truncated file.
	if err := checkFreeDiskSpace(exeDir, requiredDownloadBytes(expectedSize, exePath)); err != nil {
		logger.Printf("[UPDATE] Error: %v\n", err)
		appendUpdateLog(exePath, fmt.Sprintf("Update failed: %v", err))
		exitUpdate(1)
	}

	url := "https://github.com/Kaia-Alenia/Axon/releases/latest/download/" + binaryName
	logger.Printf("[UPDATE] Downloading latest version from:\n  %s\n", url)
	if _, err := downloadUpdateExpected(url, tmpPath, expectedSize); err != nil {
		logger.Printf("[UPDATE] Error: %v\n", err)
		appendUpdateLog(exePath, fmt.Sprintf("Update failed: %v", err))
		_ = os.Remove(tmpPath)
		exitUpdate(1)
	}
	if err := validateDownloadedBinary(tmpPath, goos); err != nil {
		logger.Printf("[UPDATE] Error: downloaded file failed validation: %v\n", err)
		appendUpdateLog(exePath, fmt.Sprintf("Update failed: downloaded file failed validation: %v", err))
		_ = os.Remove(tmpPath)
		exitUpdate(1)
	}
	// Defence in depth: the download already checked the byte count,
	// but re-check the file on disk before it is ever swapped in.
	if err := verifyDownloadedSize(tmpPath, expectedSize); err != nil {
		logger.Printf("[UPDATE] Error: %v\n", err)
		appendUpdateLog(exePath, fmt.Sprintf("Update failed: %v", err))
		_ = os.Remove(tmpPath)
		exitUpdate(1)
	}
	// Backup-first swap: the old binary is moved to <exe>.old before
	// the new one takes its place, on every platform. If the swap or
	// the launch check below fails, the old binary can be restored.
	if err := swapViaRenameAway(exePath, tmpPath); err == nil {
		if err := verifyExeVersion(exePath, latest); err != nil {
			logger.Printf("[UPDATE] Error: %v\n", err)
			if restoreErr := restoreFromBackup(exePath); restoreErr != nil {
				combined := fmt.Errorf("%v; rollback also failed: %v", err, restoreErr)
				appendUpdateLog(exePath, fmt.Sprintf("Update failed: %v", combined))
				logger.Printf("[UPDATE] Error: %v\n", combined)
			} else {
				appendUpdateLog(exePath, fmt.Sprintf("Update failed: %v; old version restored from backup", err))
				logger.Println("[UPDATE] Old version restored from backup.")
			}
			exitUpdate(1)
		}
		// The new binary is verified: only now is the backup disposable.
		if err := os.Remove(exePath + ".old"); err != nil && !os.IsNotExist(err) {
			logger.Printf("[UPDATE] Warning: could not remove backup %s: %v\n", exePath+".old", err)
		}
		logger.Println("[UPDATE] Update applied successfully. Restarting into the new version...")
		if err := restartIntoExe(exePath); err != nil {
			logger.Printf("[UPDATE] Update applied but could not restart automatically: %v\n", err)
			logger.Println("[UPDATE] Restart axon manually to use the new version.")
		}
		exitUpdate(0)
	} else {
		appendUpdateLog(exePath, fmt.Sprintf("In-process swap failed (%v); handing the swap to a helper process.", err))
	}
	// Slow path: the in-process swap failed (the running exe is
	// locked, e.g. by antivirus on Windows), so hand the swap to a
	// detached helper process and exit immediately to unlock it. The
	// arguments are passed as separate argv entries — never joined
	// into a command string — so install paths with spaces work.
	logger.Printf("[UPDATE] Could not replace the running binary (file may be in use); handing the swap to a helper process.\n")
	helperArgs := []string{"--apply-update", strconv.Itoa(os.Getpid()), exePath, tmpPath}
	if latest != "" {
		helperArgs = append(helperArgs, latest)
	}
	helper := exec.Command(exePath, helperArgs...)
	helper.Stdout = os.Stdout
	helper.Stderr = os.Stderr
	helper.Stdin = os.Stdin
	if err := helper.Start(); err != nil {
		logger.Printf("[UPDATE] Error: could not start update helper: %v\n", err)
		appendUpdateLog(exePath, fmt.Sprintf("Update failed: could not start update helper: %v", err))
		logger.Printf("[UPDATE] New binary saved at: %s\n", tmpPath)
		logger.Printf("[UPDATE] Manually replace the current binary with the downloaded file.\n")
		exitUpdate(1)
	}
	logger.Printf("[UPDATE] Update downloaded. The helper will replace the binary and restart axon once this process exits.\n")
	exitUpdate(0)
}

func main() {
	// Hidden internal flag used by the self-update helper
	// (see applyUpdate). It is intercepted here, before flag.Parse,
	// so it never appears in --help output and never starts the server.
	// Usage: axon --apply-update <parent-pid> <exe-path> <new-path> [expected-version]
	if len(os.Args) > 1 && os.Args[1] == "--apply-update" {
		if len(os.Args) != 5 && len(os.Args) != 6 {
			logger.Printf("[UPDATE] Error: --apply-update needs <parent-pid> <exe-path> <new-path> [expected-version]\n")
			exitUpdate(1)
		}
		parentPID, err := strconv.Atoi(os.Args[2])
		if err != nil {
			logger.Printf("[UPDATE] Error: invalid parent pid %q: %v\n", os.Args[2], err)
			exitUpdate(1)
		}
		expectedVersion := ""
		if len(os.Args) == 6 {
			expectedVersion = os.Args[5]
		}
		if err := applyUpdateExpected(parentPID, os.Args[3], os.Args[4], expectedVersion); err != nil {
			logger.Printf("[UPDATE] Error: %v\n", err)
			logger.Printf("[UPDATE] See update.log next to the executable for details.\n")
			exitUpdate(1)
		}
		exitUpdate(0)
	}
	// Clean up the renamed-aside binary left by a previous update
	// (see swapViaRenameAway), but only once the current binary has
	// proven it launches — never delete a backup a failed or
	// incomplete update might still need to roll back to.
	// Skipped for a "--version" probe: cleanupStaleBackup verifies by
	// spawning this same binary with --version, and that child would
	// otherwise re-enter the cleanup and recurse.
	if exePath, err := os.Executable(); err == nil {
		if len(os.Args) < 2 || os.Args[1] != "--version" {
			cleanupStaleBackup(exePath)
		}
	}
	showVersion := flag.Bool("version", false, "Show version information")
	runUpdate := flag.Bool("update", false, "Download and install the latest version of Axon")
	noLogo := flag.Bool("no-logo", false, "Skip the decorative logo (use if your terminal font lacks braille support)")
	minimized := flag.Bool("minimized", false, "Start Axon with the console window minimized")
	flag.IntVar(&port, "port", 6969, "TCP/WebSocket port to listen on (0 for auto-allocation)")
	flag.IntVar(&udpPort, "udp-port", 6970, "UDP port for fast input protocol (0 for auto-allocation)")
	flag.Parse()
	if *showVersion {
		fmt.Printf("Axon Server v%s\n", version)
		fmt.Println("High-speed wireless input bridge")
		exitUpdate(0)
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
		exitUpdate(1)
	}
	tcpListener, actualPort, err := tryPortWithFallback(port)
	if err != nil {
		logger.Printf(" Error: Failed to bind to any port: %v\n", err)
		exitUpdate(1)
	}
	defer tcpListener.Close()
	if actualPort != port {
		logger.Printf("[PORT] Using dynamic port: %d (requested: %d)\n", actualPort, port)
		port = actualPort
	}

	udpConn, actualUdpPort, err := tryUDPPortWithFallback(udpPort)
	if err != nil {
		logger.Printf(" Error: Failed to bind UDP port: %v\n", err)
		exitUpdate(1)
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
		exitUpdate(1)
	}
	http.Handle("/", http.FileServer(http.FS(subFS)))
	http.HandleFunc("/ws", handleWebSocket)
	server := &http.Server{}
	logger.Printf("[SERVER]  Listening on port %d...\n", port)
	err = server.Serve(tcpListener)
	if err != nil {
		logger.Printf(" Error starting the server: %v\n", err)
		exitUpdate(1)
	}
}
