//go:build !windows

package main

import (
	"syscall"
)

func setReuseAddrControl(network, address string, c syscall.RawConn) error {
	var opErr error
	err := c.Control(func(fd uintptr) {
		opErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})
	if opErr != nil {
		return opErr
	}
	return err
}
