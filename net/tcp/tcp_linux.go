//go:build linux

package tcp

import "golang.org/x/sys/unix"

// bindToDev sets the SO_BINDTODEVICE option
func bindToDev(fd int, devName string) error {
	return unix.SetsockoptString(fd, unix.IPPROTO_TCP, unix.SO_BINDTODEVICE, devName)
}
