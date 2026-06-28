//go:build windows

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

var bluetoothListenerWindows net.Listener

func startBluetoothServer() {
	// Windows Bluetooth RFCOMM support via Winsock2
	// Note: Full Windows Bluetooth API support requires CGo with Windows Bluetooth headers
	// For now, we provide a TCP-based listener for Bluetooth client redirection
	
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("[BT] Error creating Bluetooth listener: %v\n", err)
		return
	}
	defer listener.Close()
	bluetoothListenerWindows = listener

	fmt.Printf("RFCOMM Bluetooth server listening on %s (Windows)\n", listener.Addr().String())

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

		// Optimize TCP connection for low-latency
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			remoteAddr := tcpConn.RemoteAddr().String()
			fmt.Printf("[CONNECTION] Bluetooth client connected from: %s\n", remoteAddr)
			
			// Set TCP options for low-latency
			_ = tcpConn.SetNoDelay(true)
			_ = tcpConn.SetKeepAlive(true)
			_ = tcpConn.SetKeepAlivePeriod(1 * time.Second)
		} else {
			fmt.Println("[CONNECTION] Bluetooth client connected")
		}

		go handleBluetoothConnection(conn)
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
