package tcp

import (
	"net"
	"syscall"
	"unsafe"
)

const (
	tcpMD5SIG           = 14 // (RFC2385)
	tcpMD5SIGMaxKeyLen  = 80
	tcpMD5SIGFlagPrefix = 0
)

type tcpMD5sig struct {
	ssFamily  uint16
	ss        [126]byte
	flags     uint8
	prefixLen uint8
	keylen    uint16
	ifIndex   uint32
	key       [tcpMD5SIGMaxKeyLen]byte
}

func buildTCPMD5Sig(addr net.IP, key string) tcpMD5sig {
	family := syscall.AF_INET
	if addr.To4() == nil {
		family = syscall.AF_INET6
	}

	t := tcpMD5sig{
		ssFamily:  uint16(family),
		flags:     tcpMD5SIGFlagPrefix,
		prefixLen: 0,
		keylen:    uint16(len(key)),
	}

	if family == syscall.AF_INET {
		copy(t.ss[2:], addr.To4())
	} else {
		copy(t.ss[2:], addr.To16())
	}

	copy(t.key[0:], []byte(key))

	return t
}

func setTCPMD5Option(fd int, addr net.IP, md5secret string) error {
	sig := buildTCPMD5Sig(addr, md5secret)
	b := *(*[unsafe.Sizeof(sig)]byte)(unsafe.Pointer(&sig))
	return syscall.SetsockoptString(fd, syscall.IPPROTO_TCP, tcpMD5SIG, string(b[:]))
}
