//go:build linux

package tcp

import (
	"net"

	"golang.org/x/sys/unix"
)

const (
	tcpMD5SIGFlagPrefix = 0
)

func buildTCPMD5Sig(addr net.IP, key string) *unix.TCPMD5Sig {
	t := &unix.TCPMD5Sig{
		Flags:     tcpMD5SIGFlagPrefix,
		Prefixlen: 0,
		Keylen:    uint16(len(key)),
	}

	if addr.To4() != nil {
		t.Addr.Family = unix.AF_INET
		copy(t.Addr.Data[2:], addr.To4())
	} else {
		t.Addr.Family = unix.AF_INET6
		copy(t.Addr.Data[6:], addr.To16())
	}

	copy(t.Key[0:], key)

	return t
}

func setTCPMD5Option(fd int, addr net.IP, md5secret string) error {
	sig := buildTCPMD5Sig(addr, md5secret)
	return unix.SetsockoptTCPMD5Sig(fd, unix.IPPROTO_TCP, unix.TCP_MD5SIG, sig)
}
