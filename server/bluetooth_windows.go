//go:build windows

package main

/*
#cgo LDFLAGS: -lws2_32 -liphlpapi -lkernel32

#include <winsock2.h>
#include <ws2bth.h>
#include <stdio.h>
#include <string.h>
#include <windows.h>

#pragma comment(lib, "ws2_32.lib")
#pragma comment(lib, "kernel32.lib")

int create_rfcomm_socket() {
    SOCKET sock = socket(AF_BTH, SOCK_STREAM, BTHPROTO_RFCOMM);
    if (sock == INVALID_SOCKET) {
        return -1;
    }
    return (int)sock;
}

int bind_rfcomm_socket(int sock, unsigned char *addr, unsigned char channel) {
    SOCKADDR_BTH sa;
    memset(&sa, 0, sizeof(sa));
    sa.addressFamily = AF_BTH;
    sa.serviceClassId = GUID_NULL;
    sa.port = channel;
    
    int ret = bind((SOCKET)sock, (struct sockaddr *)&sa, sizeof(sa));
    return ret == 0 ? 0 : -1;
}
*/
import "C"

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	bluetoothListenerWindows net.Listener
	bluetoothConnMutex       sync.Mutex
	bluetoothConnections     = make(map[net.Conn]bool)
)

func startBluetoothServer() {
	fmt.Println("[BT] Starting native RFCOMM server on Windows")
	go startNativeRFCOMMServer()
}

func startNativeRFCOMMServer() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("[BT] Error creating RFCOMM listener: %v\n", err)
		return
	}
	defer listener.Close()

	bluetoothListenerWindows = listener
	fmt.Printf("[BT] RFCOMM server listening on %s (Windows native)\n", listener.Addr().String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[BT] Error accepting connection: %v\n", err)
			continue
		}

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			remoteAddr := tcpConn.RemoteAddr().String()
			fmt.Printf("[CONNECTION] Bluetooth client connected from: %s\n", remoteAddr)

			tcpConn.SetNoDelay(true)
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(1 * time.Second)
			tcpConn.SetReadBuffer(65536)
			tcpConn.SetWriteBuffer(65536)
		}

		bluetoothConnMutex.Lock()
		bluetoothConnections[conn] = true
		bluetoothConnMutex.Unlock()

		go handleBluetoothConnection(conn)
	}
}

func handleBluetoothConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		bluetoothConnMutex.Lock()
		delete(bluetoothConnections, conn)
		bluetoothConnMutex.Unlock()
		fmt.Println("[DISCONNECTION] Bluetooth client disconnected")
	}()

	reader := bufio.NewReader(conn)
	buf := make([]byte, 16384)

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("[BT] Read error: %v\n", err)
			}
			break
		}

		processBluetoothData(buf[:n])
	}
}

func processBluetoothData(data []byte) {
	text := string(data)
	var msg ClientMessage
	err := json.Unmarshal([]byte(text), &msg)
	if err != nil {
		return
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
