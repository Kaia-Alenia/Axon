//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework IOBluetooth -framework Foundation

#import <IOBluetooth/IOBluetooth.h>
#import <Foundation/Foundation.h>

typedef int (*rfcomm_callback_t)(int channel, const uint8_t* data, int len);

static rfcomm_callback_t g_callback = NULL;
static IOBluetoothRFCOMMChannel* g_rfcomm_channel = NULL;

void set_rfcomm_callback(rfcomm_callback_t callback) {
    g_callback = callback;
}

int start_rfcomm_server() {
    // Create an RFCOMM service on a dynamic channel
    IOBluetoothSDPServiceRecord* serviceRecord;
    NSError* error = nil;
    
    NSMutableDictionary* attributes = [NSMutableDictionary dictionary];
    [attributes setObject:@"Axon RFCOMM Server" forKey:@"ServiceName"];
    [attributes setObject:@"Axon High-Speed Input" forKey:@"ServiceDescription"];
    
    // Publish service on dynamic RFCOMM channel
    IOReturn ret = IOBluetoothAddServiceDict(
        (CFDictionaryRef)attributes,
        &serviceRecord
    );
    
    if (ret != kIOReturnSuccess) {
        return -1;
    }
    
    return 1;
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
	bluetoothListenerDarwin net.Listener
	bluetoothConnMutex      sync.Mutex
	bluetoothConnections    = make(map[net.Conn]bool)
)

func startBluetoothServer() {
	fmt.Println("[BT] Starting native RFCOMM server on macOS")
	go startNativeRFCOMMServer()
}

func startNativeRFCOMMServer() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("[BT] Error creating RFCOMM listener: %v\n", err)
		return
	}
	defer listener.Close()

	bluetoothListenerDarwin = listener
	fmt.Printf("[BT] RFCOMM server listening on %s (macOS native)\n", listener.Addr().String())

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
	lines := bufio.NewScanner(bufio.NewReader(nil))
	lines.Buffer(data, len(data))

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
