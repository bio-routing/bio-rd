package tcp

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

// Listener listens for TCP clients
type Listener struct {
	fd    int
	laddr *net.TCPAddr
}

// Listen starts a TCPListener
func Listen(laddr *net.TCPAddr, ttl uint8) (*Listener, error) {
	l := &Listener{
		laddr: laddr,
	}

	afi := unix.AF_INET
	if laddr.IP.To4() == nil {
		afi = unix.AF_INET6
	}

	fd, err := unix.Socket(afi, unix.SOCK_STREAM, unix.IPPROTO_TCP)
	if err != nil {
		return nil, fmt.Errorf("socket() failed: %w", err)
	}
	l.fd = fd

	if afi == unix.AF_INET6 {
		err = unix.SetsockoptInt(fd, SOL_IPV6, unix.IPV6_V6ONLY, 1)
		if err != nil {
			unix.Close(fd)
			return nil, fmt.Errorf("unable to set IPV6_V6ONLY: %w", err)
		}
	}

	err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("unable to get SO_REUSEADDR %w", err)
	}

	if ttl != 0 {
		err = unix.SetsockoptInt(fd, SOL_IP, unix.IP_TTL, int(ttl))
		if err != nil {
			unix.Close(fd)
			return nil, fmt.Errorf("unable to set IP_TTL: %w", err)
		}
	}

	if laddr.IP.To4() != nil {
		err = unix.Bind(fd, &unix.SockaddrInet4{
			Port: laddr.Port,
			Addr: ipv4AddrToArray(laddr.IP),
		})
	} else {
		err = unix.Bind(fd, &unix.SockaddrInet6{
			Port: laddr.Port,
			Addr: ipv6AddrToArray(laddr.IP),
		})
	}
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("bind failed: %w", err)
	}

	err = unix.Listen(fd, 128)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("listen failed: %w", err)
	}

	return l, nil
}

// SetTCPMD5 sets a TCP md5 secret for addr
func (l *Listener) SetTCPMD5(peerAddr net.IP, secret string) error {
	isIPv4Listener := l.laddr.IP.To4() != nil
	isIPv4Client := peerAddr.To4() != nil

	// Do not try to set MD5 secret if listener and peerAddr are of different AFIs.
	// Call to setsockopt() would fail with -EINVAL. This is also why we use separate listeners
	// per AFI. Tested for you by takt
	if isIPv4Client != isIPv4Listener {
		return nil
	}

	return setTCPMD5Option(l.fd, peerAddr, secret)
}

// AcceptTCP accepts a new TCP connection
func (l *Listener) AcceptTCP() (*Conn, error) {
	fd, sa, err := unix.Accept(l.fd)
	if err != nil {
		return nil, err
	}

	raddr := &net.TCPAddr{
		Port: 0,
	}

	switch sa.(type) {
	case *unix.SockaddrInet4:
		x := sa.(*unix.SockaddrInet4)
		raddr.IP = x.Addr[:]
		raddr.Port = x.Port
	case *unix.SockaddrInet6:
		x := sa.(*unix.SockaddrInet4)
		raddr.IP = x.Addr[:]
		raddr.Port = x.Port
	}

	return &Conn{
		fd:    fd,
		laddr: l.laddr,
		raddr: raddr,
	}, nil
}
