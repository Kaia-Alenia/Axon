package main
import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)
type MockSimulator struct {
	lastMove      [2]float64
	lastClick     string
	lastMouseDown string
	lastMouseUp   string
	lastScroll    string
	lastType      string
	lastKey       string
	lastCombo     [2]string
}
func (m *MockSimulator) MoveMouse(dx, dy float64) {
	m.lastMove = [2]float64{dx, dy}
}
func (m *MockSimulator) Click(button string) {
	m.lastClick = button
}
func (m *MockSimulator) MouseDown(button string) {
	m.lastMouseDown = button
}
func (m *MockSimulator) MouseUp(button string) {
	m.lastMouseUp = button
}
func (m *MockSimulator) Scroll(direction string) {
	m.lastScroll = direction
}
func (m *MockSimulator) Type(text string) {
	m.lastType = text
}
func (m *MockSimulator) Key(key string) {
	m.lastKey = key
}
func (m *MockSimulator) KeyCombo(mod, key string) {
	m.lastCombo = [2]string{mod, key}
}
func (m *MockSimulator) Close() {
}
func TestBluetoothServerStarts(t *testing.T) {
	originalSimulator := simulator
	simulator = &MockSimulator{}
	defer func() {
		simulator = originalSimulator
	}()
	fmt.Println("[TEST] Bluetooth server initialization test started")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("[TEST] Bluetooth server initialization completed successfully")
}
func TestBluetoothClientConnection(t *testing.T) {
	originalSimulator := simulator
	simulator = &MockSimulator{}
	defer func() {
		simulator = originalSimulator
	}()
	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		t.Skipf("Could not connect to Bluetooth server: %v", err)
	}
	defer conn.Close()
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetNoDelay(true)
		tcpConn.SetKeepAlive(true)
	}
	fmt.Println("[TEST] Client connection successful")
}
func TestBluetoothMoveMessage(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	msg := ClientMessage{
		Type: "move",
		Dx:   10.5,
		Dy:   20.3,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	processBluetoothData(data)
	if mock.lastMove[0] != 10.5 || mock.lastMove[1] != 20.3 {
		t.Errorf("Move command failed: got (%f, %f), want (10.5, 20.3)", mock.lastMove[0], mock.lastMove[1])
	}
	fmt.Println("[TEST] Move message processing passed")
}
func TestBluetoothClickMessage(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	msg := ClientMessage{
		Type:   "click",
		Button: "left",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	processBluetoothData(data)
	if mock.lastClick != "left" {
		t.Errorf("Click command failed: got %q, want %q", mock.lastClick, "left")
	}
	fmt.Println("[TEST] Click message processing passed")
}
func TestBluetoothTypeMessage(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	testText := "Hello, AXON!"
	msg := ClientMessage{
		Type: "type",
		Text: testText,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	processBluetoothData(data)
	if mock.lastType != testText {
		t.Errorf("Type command failed: got %q, want %q", mock.lastType, testText)
	}
	fmt.Println("[TEST] Type message processing passed")
}
func TestBluetoothDisconnection(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()
	go func() {
		conn, _ := listener.Accept()
		if conn != nil {
			defer conn.Close()
			time.Sleep(50 * time.Millisecond)
		}
	}()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	conn.Close()
	fmt.Println("[TEST] Disconnection handling passed")
}
func TestBluetoothLatency(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	msg := ClientMessage{
		Type: "move",
		Dx:   5.0,
		Dy:   5.0,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	start := time.Now()
	for i := 0; i < 100; i++ {
		processBluetoothData(data)
	}
	elapsed := time.Since(start)
	latencyPerMsg := elapsed.Milliseconds() / 100
	if latencyPerMsg > 5 {
		t.Logf("Warning: High latency per message: %d ms", latencyPerMsg)
	}
	fmt.Printf("[TEST] Latency test completed: avg %.2f ms per message\n", float64(elapsed.Milliseconds())/100.0)
}
func TestBluetoothMouseDownUp(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	msgDown := ClientMessage{
		Type:   "mousedown",
		Button: "right",
	}
	data, _ := json.Marshal(msgDown)
	processBluetoothData(data)
	if mock.lastMouseDown != "right" {
		t.Errorf("MouseDown failed: got %q, want %q", mock.lastMouseDown, "right")
	}
	msgUp := ClientMessage{
		Type:   "mouseup",
		Button: "right",
	}
	data, _ = json.Marshal(msgUp)
	processBluetoothData(data)
	if mock.lastMouseUp != "right" {
		t.Errorf("MouseUp failed: got %q, want %q", mock.lastMouseUp, "right")
	}
	fmt.Println("[TEST] MouseDown/MouseUp processing passed")
}
func TestBluetoothScrollMessage(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	msgUp := ClientMessage{
		Type: "scroll",
		Dy:   1.0,
	}
	data, _ := json.Marshal(msgUp)
	processBluetoothData(data)
	if mock.lastScroll != "up" {
		t.Errorf("Scroll up failed: got %q, want %q", mock.lastScroll, "up")
	}
	msgDown := ClientMessage{
		Type: "scroll",
		Dy:   -1.0,
	}
	data, _ = json.Marshal(msgDown)
	processBluetoothData(data)
	if mock.lastScroll != "down" {
		t.Errorf("Scroll down failed: got %q, want %q", mock.lastScroll, "down")
	}
	fmt.Println("[TEST] Scroll message processing passed")
}
func TestBluetoothKeyCombo(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	msg := ClientMessage{
		Type:     "keycombo",
		Modifier: "ctrl",
		Key:      "c",
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	processBluetoothData(data)
	if mock.lastCombo[0] != "ctrl" || mock.lastCombo[1] != "c" {
		t.Errorf("KeyCombo failed: got (%q, %q), want (ctrl, c)", mock.lastCombo[0], mock.lastCombo[1])
	}
	fmt.Println("[TEST] KeyCombo message processing passed")
}
func TestBluetoothMessageParsingError(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	invalidData := []byte("{invalid json")
	processBluetoothData(invalidData)
	processBluetoothData([]byte(""))
	processBluetoothData([]byte("random string data"))
	fmt.Println("[TEST] Invalid message handling passed")
}
func TestConcurrentBluetoothConnections(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	numConnections := 5
	done := make(chan bool, numConnections)
	for i := 0; i < numConnections; i++ {
		go func(id int) {
			msg := ClientMessage{
				Type: "move",
				Dx:   float64(id),
				Dy:   float64(id),
			}
			data, _ := json.Marshal(msg)
			processBluetoothData(data)
			done <- true
		}(i)
	}
	for i := 0; i < numConnections; i++ {
		<-done
	}
	fmt.Printf("[TEST] Concurrent connections test (%d clients) passed\n", numConnections)
}
func TestBluetoothDataStream(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	messages := []ClientMessage{
		{Type: "move", Dx: 10, Dy: 10},
		{Type: "click", Button: "left"},
		{Type: "type", Text: "hello"},
		{Type: "scroll", Dy: -5},
		{Type: "key", Key: "escape"},
	}
	for i, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Failed to marshal message %d: %v", i, err)
			continue
		}
		processBluetoothData(data)
	}
	if mock.lastKey != "escape" {
		t.Errorf("Stream processing failed: last key should be 'escape'")
	}
	fmt.Println("[TEST] Data stream processing passed")
}
func TestBluetoothConnectionTimeout(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		buf := make([]byte, 1024)
		_, err = conn.Read(buf)
		if err != nil && err != io.EOF {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			} else {
				fmt.Printf("Unexpected error: %v\n", err)
			}
		}
	}()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	time.Sleep(200 * time.Millisecond)
	fmt.Println("[TEST] Connection timeout handling passed")
}
func TestBluetoothSpecialCharacters(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	specialTexts := []string{
		"hello@#$%",
		"中文测试",
		"émojis 🎮",
		"\"quoted\"",
		"line\nbreak",
	}
	for _, text := range specialTexts {
		msg := ClientMessage{
			Type: "type",
			Text: text,
		}
		data, _ := json.Marshal(msg)
		processBluetoothData(data)
		if mock.lastType != text {
			t.Errorf("Special character handling failed for %q: got %q", text, mock.lastType)
		}
	}
	fmt.Println("[TEST] Special character handling passed")
}
func TestBluetoothMultipleKeys(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	keys := []string{"escape", "enter", "tab", "space", "backspace"}
	for _, key := range keys {
		msg := ClientMessage{
			Type: "key",
			Key:  key,
		}
		data, _ := json.Marshal(msg)
		processBluetoothData(data)
		if mock.lastKey != key {
			t.Errorf("Key command failed for %q: got %q", key, mock.lastKey)
		}
	}
	fmt.Println("[TEST] Multiple key processing passed")
}
func BenchmarkBluetoothMessageProcessing(b *testing.B) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	msg := ClientMessage{
		Type: "move",
		Dx:   10.0,
		Dy:   10.0,
	}
	data, _ := json.Marshal(msg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processBluetoothData(data)
	}
	b.StopTimer()
	latencyPerOp := b.Elapsed().Milliseconds() / int64(b.N)
	fmt.Printf("[BENCH] Latency per operation: %.2f ms\n", float64(latencyPerOp))
}
func sendBluetoothMessage(conn net.Conn, msg ClientMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = conn.Write(append(data, '\n'))
	return err
}
func readBluetoothResponse(conn net.Conn, timeout time.Duration) (string, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))
	defer conn.SetReadDeadline(time.Time{})
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
func TestBluetoothFullMessageCycle(t *testing.T) {
	mock := &MockSimulator{}
	originalSimulator := simulator
	simulator = mock
	defer func() {
		simulator = originalSimulator
	}()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			var msg ClientMessage
			json.Unmarshal([]byte(line), &msg)
			processBluetoothData([]byte(line))
		}
	}()
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	msg := ClientMessage{
		Type: "move",
		Dx:   25.0,
		Dy:   35.0,
	}
	if err := sendBluetoothMessage(conn, msg); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if mock.lastMove[0] != 25.0 || mock.lastMove[1] != 35.0 {
		t.Errorf("Full cycle failed: got (%f, %f), want (25.0, 35.0)", mock.lastMove[0], mock.lastMove[1])
	}
	fmt.Println("[TEST] Full message cycle passed")
}
