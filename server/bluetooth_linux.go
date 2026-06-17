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
		fmt.Println("Error al crear socket Bluetooth:", err)
		return
	}
	defer unix.Close(fd)

	addr := &unix.SockaddrRFCOMM{
		Channel: 1,
		Addr:    [6]uint8{0, 0, 0, 0, 0, 0}, // Any address
	}

	if err := unix.Bind(fd, addr); err != nil {
		fmt.Println("Error al hacer bind en Bluetooth (quizás otro servicio usa el canal 1):", err)
		return
	}

	if err := unix.Listen(fd, 1); err != nil {
		fmt.Println("Error al escuchar en Bluetooth:", err)
		return
	}

	fmt.Println("Servidor Bluetooth RFCOMM escuchando en el canal 1")

	for {
		clientFd, sa, err := unix.Accept(fd)
		if err != nil {
			fmt.Println("Error al aceptar conexión Bluetooth:", err)
			continue
		}

		if btAddr, ok := sa.(*unix.SockaddrRFCOMM); ok {
			fmt.Printf("[CONEXIÓN] Cliente Bluetooth conectado desde MAC: %02X:%02X:%02X:%02X:%02X:%02X\n",
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
		
		var msg ClientMessage
		err := json.Unmarshal([]byte(text), &msg)
		if err != nil {
			// Handle plain strings if needed, or ignore
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
			fmt.Printf("[TECLADO BT] Escribiendo texto: %q\n", msg.Text)
			simulator.Type(msg.Text)
		case "key":
			fmt.Printf("[TECLADO BT] Presionando tecla especial: %q\n", msg.Key)
			simulator.Key(msg.Key)
		case "keycombo":
			fmt.Printf("[TECLADO BT] Presionando combo: %s+%s\n", msg.Modifier, msg.Key)
			simulator.KeyCombo(msg.Modifier, msg.Key)
		}
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Println("[DESCONEXIÓN] Cliente Bluetooth error:", err)
	} else {
		fmt.Println("[DESCONEXIÓN] Cliente Bluetooth desconectado")
	}
}
