//go:build !linux

package tcp

import (
	"fmt"
	"net"
)

func setTCPMD5Option(fd int, addr net.IP, md5secret string) error {
	return fmt.Errorf("setting md5 is not supported")
}
