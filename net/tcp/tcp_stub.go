//go:build !linux

package tcp

import "fmt"

// bindToDev sets the SO_BINDTODEVICE option
func bindToDev(fd int, devName string) error {
	return fmt.Errorf("binding to device is not supported")
}
