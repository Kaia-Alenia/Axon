//go:build windows

package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"golang.org/x/sys/windows"
	"io"
	"unsafe"
)

const (
	afBth          = 32
	bthprotoRfcomm = 3
	btPortAny      = 0xffffffff
	nsBth          = 16
	rnrRegister    = 0
)

var (
	ws2_32            = windows.NewLazySystemDLL("ws2_32.dll")
	procSocket        = ws2_32.NewProc("socket")
	procBind          = ws2_32.NewProc("bind")
	procListen        = ws2_32.NewProc("listen")
	procAccept        = ws2_32.NewProc("accept")
	procGetsockname   = ws2_32.NewProc("getsockname")
	procWSASetService = ws2_32.NewProc("WSASetServiceW")
	procRecv          = ws2_32.NewProc("recv")
	procClosesocket   = ws2_32.NewProc("closesocket")
	sppServiceGUID    = windows.GUID{
		Data1: 0x00001101,
		Data2: 0x0000,
		Data3: 0x1000,
		Data4: [8]byte{0x80, 0x00, 0x00, 0x80, 0x5F, 0x9B, 0x34, 0xFB},
	}
)

type btSocketAddr struct {
	Sockaddr       uintptr
	SockaddrLength int32
}

type btCsAddrInfo struct {
	LocalAddr  btSocketAddr
	RemoteAddr btSocketAddr
	SocketType int32
	Protocol   int32
}

type btWsaQuerySetW struct {
	Size                uint32
	ServiceInstanceName *uint16
	ServiceClassId      *windows.GUID
	Version             uintptr
	Comment             *uint16
	NameSpace           uint32
	NSProviderId        *windows.GUID
	Context             *uint16
	NumberOfProtocols   uint32
	AfpProtocols        uintptr
	QueryString         *uint16
	NumberOfCsAddrs     uint32
	SaBuffer            *btCsAddrInfo
	OutputFlags         uint32
	Blob                uintptr
}

type rfcommConn struct {
	handle uintptr
}

func (c *rfcommConn) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	r1, _, err := procRecv.Call(
		c.handle,
		uintptr(unsafe.Pointer(&p[0])),
		uintptr(len(p)),
		0,
	)
	n := int(int32(r1))
	if n == 0 {
		return 0, io.EOF
	}
	if n < 0 {
		return 0, err
	}
	return n, nil
}

func (c *rfcommConn) Close() error {
	procClosesocket.Call(c.handle)
	return nil
}

func makeSockaddrBth(port uint32) [30]byte {
	var sa [30]byte
	binary.LittleEndian.PutUint16(sa[0:2], afBth)
	binary.LittleEndian.PutUint32(sa[10:14], sppServiceGUID.Data1)
	binary.LittleEndian.PutUint16(sa[14:16], sppServiceGUID.Data2)
	binary.LittleEndian.PutUint16(sa[16:18], sppServiceGUID.Data3)
	copy(sa[18:26], sppServiceGUID.Data4[:])
	binary.LittleEndian.PutUint32(sa[26:30], port)
	return sa
}

func startBluetoothServer() {
	var wsaData windows.WSAData
	if err := windows.WSAStartup(uint32(0x0202), &wsaData); err != nil {
		logger.Printf("[BT] WSAStartup failed: %v\n", err)
		return
	}
	sock, _, err := procSocket.Call(afBth, windows.SOCK_STREAM, bthprotoRfcomm)
	if sock == ^uintptr(0) {
		logger.Printf("[BT] Error creating Bluetooth RFCOMM socket: %v\n", err)
		return
	}
	defer procClosesocket.Call(sock)
	sa := makeSockaddrBth(1)
	r1, _, _ := procBind.Call(sock, uintptr(unsafe.Pointer(&sa[0])), 30)
	if int32(r1) != 0 {
		sa = makeSockaddrBth(btPortAny)
		r1, _, err = procBind.Call(sock, uintptr(unsafe.Pointer(&sa[0])), 30)
		if int32(r1) != 0 {
			logger.Printf("[BT] Error binding RFCOMM socket: %v\n", err)
			return
		}
	}
	r1, _, err = procListen.Call(sock, 4)
	if int32(r1) != 0 {
		logger.Printf("[BT] Error listening on RFCOMM socket: %v\n", err)
		return
	}
	nameLen := int32(30)
	procGetsockname.Call(sock, uintptr(unsafe.Pointer(&sa[0])), uintptr(unsafe.Pointer(&nameLen)))
	boundChannel := binary.LittleEndian.Uint32(sa[26:30])
	serviceName, _ := windows.UTF16PtrFromString("Axon Remote Input")
	csAddr := btCsAddrInfo{
		LocalAddr: btSocketAddr{
			Sockaddr:       uintptr(unsafe.Pointer(&sa[0])),
			SockaddrLength: 30,
		},
		RemoteAddr: btSocketAddr{
			Sockaddr:       uintptr(unsafe.Pointer(&sa[0])),
			SockaddrLength: 30,
		},
		SocketType: windows.SOCK_STREAM,
		Protocol:   bthprotoRfcomm,
	}
	qs := btWsaQuerySetW{
		Size:                uint32(unsafe.Sizeof(btWsaQuerySetW{})),
		ServiceInstanceName: serviceName,
		ServiceClassId:      &sppServiceGUID,
		NameSpace:           nsBth,
		NumberOfCsAddrs:     1,
		SaBuffer:            &csAddr,
	}
	procWSASetService.Call(uintptr(unsafe.Pointer(&qs)), rnrRegister, 0)
	logger.Printf("[BT] RFCOMM Bluetooth server listening on channel %d\n", boundChannel)
	for {
		var clientSa [30]byte
		clientSaLen := int32(30)
		clientSock, _, err := procAccept.Call(
			sock,
			uintptr(unsafe.Pointer(&clientSa[0])),
			uintptr(unsafe.Pointer(&clientSaLen)),
		)
		if clientSock == ^uintptr(0) {
			logger.Printf("[BT] Error accepting Bluetooth connection: %v\n", err)
			continue
		}
		logger.Printf("[CONNECTION] Bluetooth client connected from MAC: %02X:%02X:%02X:%02X:%02X:%02X\n",
			clientSa[7], clientSa[6], clientSa[5], clientSa[4], clientSa[3], clientSa[2])
		go handleBluetoothConnection(&rfcommConn{handle: clientSock})
	}
}

func handleBluetoothConnection(conn io.ReadCloser) {
	defer func() {
		conn.Close()
		logger.Println("[DISCONNECTION] Bluetooth client disconnected")
	}()
	scanner := bufio.NewScanner(conn)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		processBluetoothData(scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		logger.Printf("[BT] Read error: %v\n", err)
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
		logger.Printf("[BT MOUSE] Click: %s\n", msg.Button)
		simulator.Click(msg.Button)
	case "mousedown":
		logger.Printf("[BT MOUSE] MouseDown: %s\n", msg.Button)
		simulator.MouseDown(msg.Button)
	case "mouseup":
		logger.Printf("[BT MOUSE] MouseUp: %s\n", msg.Button)
		simulator.MouseUp(msg.Button)
	case "scroll":
		dir := "down"
		if msg.Dy > 0 {
			dir = "up"
		}
		simulator.Scroll(dir)
	case "type":
		logger.Printf("[BT KEYBOARD] Typing text: %q\n", msg.Text)
		simulator.Type(msg.Text)
	case "key":
		logger.Printf("[BT KEYBOARD] Pressing special key: %q\n", msg.Key)
		simulator.Key(msg.Key)
	case "keycombo":
		logger.Printf("[BT KEYBOARD] Pressing combo: %s+%s\n", msg.Modifier, msg.Key)
		simulator.KeyCombo(msg.Modifier, msg.Key)
	}
}
