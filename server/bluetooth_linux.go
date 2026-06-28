//go:build linux

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func startBluetoothServer() {
	fd, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		fmt.Println("Error creating Bluetooth socket:", err)
		return
	}
	defer unix.Close(fd)

	addr := &unix.SockaddrRFCOMM{
		Channel: 1,
		Addr:    [6]uint8{0, 0, 0, 0, 0, 0}, // Any address
	}

	if err := unix.Bind(fd, addr); err != nil {
		fmt.Println("Error binding to Bluetooth (maybe another service is using channel 1):", err)
		return
	}

	if err := unix.Listen(fd, 1); err != nil {
		fmt.Println("Error listening on Bluetooth:", err)
		return
	}

	fmt.Println("RFCOMM Bluetooth server listening on channel 1")

	for {
		clientFd, sa, err := unix.Accept(fd)
		if err != nil {
			fmt.Println("Error accepting Bluetooth connection:", err)
			continue
		}

		if btAddr, ok := sa.(*unix.SockaddrRFCOMM); ok {
			fmt.Printf("[CONNECTION] Bluetooth client connected from MAC: %02X:%02X:%02X:%02X:%02X:%02X\n",
				btAddr.Addr[5], btAddr.Addr[4], btAddr.Addr[3], btAddr.Addr[2], btAddr.Addr[1], btAddr.Addr[0])
		}

		go handleBluetoothConnection(clientFd)
	}
}

func handleBluetoothConnection(fd int) {
	file := os.NewFile(uintptr(fd), "rfcomm")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		processBluetoothData([]byte(text))
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Println("[DISCONNECTION] Bluetooth client error:", err)
	} else {
		fmt.Println("[DISCONNECTION] Bluetooth client disconnected")
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
