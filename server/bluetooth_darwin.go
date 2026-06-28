//go:build darwin

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var bluetoothListenerDarwin net.Listener

func startBluetoothServer() {
	// macOS: Use TCP-based listener since native Bluetooth socket support
	// requires IOBluetooth framework integration which needs CGo
	fmt.Println("[BT] Starting Bluetooth listener on macOS")
	startTCPBluetoothServer()
}

func startNativeBluetoothServer() bool {
	// Kept for future enhancements with IOBluetooth framework
	return false
}

func startTCPBluetoothServer() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("[BT] Error creating Bluetooth listener: %v\n", err)
		return
	}
	defer listener.Close()
	bluetoothListenerDarwin = listener

	fmt.Printf("TCP-based Bluetooth server listening on %s\n", listener.Addr().String())

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		listener.Close()
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[BT] Error accepting connection:", err)
			continue
		}

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			remoteAddr := tcpConn.RemoteAddr().String()
			fmt.Printf("[CONNECTION] Bluetooth client connected from: %s\n", remoteAddr)
			tcpConn.SetNoDelay(true)
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(1 * time.Second)
		}

		go handleBluetoothConnection(conn)
	}
}

func handleBluetoothConnectionFd(fd int) {
	file := os.NewFile(uintptr(fd), "rfcomm")
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("[DISCONNECTION] Bluetooth client error:", err)
			} else {
				fmt.Println("[DISCONNECTION] Bluetooth client disconnected")
			}
			return
		}

		var msg ClientMessage
		err = json.Unmarshal([]byte(line), &msg)
		if err != nil {
			continue
		}

		switch msg.Type {
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
			fmt.Printf("[BT KEYBOARD] Typing text: %q\n", msg.Text)
			simulator.Type(msg.Text)
		case "key":
			fmt.Printf("[BT KEYBOARD] Pressing special key: %q\n", msg.Key)
			simulator.Key(msg.Key)
		case "keycombo":
			fmt.Printf("[BT KEYBOARD] Pressing combo: %s+%s\n", msg.Modifier, msg.Key)
			simulator.KeyCombo(msg.Modifier, msg.Key)
		}
	}
}

func handleBluetoothConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("[DISCONNECTION] Bluetooth client error:", err)
			} else {
				fmt.Println("[DISCONNECTION] Bluetooth client disconnected")
			}
			return
		}

		var msg ClientMessage
		err = json.Unmarshal([]byte(line), &msg)
		if err != nil {
			continue
		}

		switch msg.Type {
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
			fmt.Printf("[BT KEYBOARD] Typing text: %q\n", msg.Text)
			simulator.Type(msg.Text)
		case "key":
			fmt.Printf("[BT KEYBOARD] Pressing special key: %q\n", msg.Key)
			simulator.Key(msg.Key)
		case "keycombo":
			fmt.Printf("[BT KEYBOARD] Pressing combo: %s+%s\n", msg.Modifier, msg.Key)
			simulator.KeyCombo(msg.Modifier, msg.Key)
		}
	}
}
