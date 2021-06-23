package tcp

import (
	"net"
	"syscall"

	"github.com/pkg/errors"
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

	afi := syscall.AF_INET
	if laddr.IP.To4() == nil {
		afi = syscall.AF_INET6
	}

	fd, err := syscall.Socket(afi, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return nil, errors.Wrap(err, "socket() failed")
	}
	l.fd = fd

	if afi == syscall.AF_INET6 {
		err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IPV6, syscall.IPV6_V6ONLY, 1)
		if err != nil {
			syscall.Close(fd)
			return nil, errors.Wrap(err, "Unable to set IPV6_V6ONLY")
		}
	}

	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		syscall.Close(fd)
		return nil, errors.Wrap(err, "Unable to get SO_REUSEADDR")
	}

	if ttl != 0 {
		err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_TTL, int(ttl))
		if err != nil {
			syscall.Close(fd)
			return nil, errors.Wrap(err, "Unable to set IP_TTL")
		}
	}

	if laddr.IP.To4() != nil {
		err = syscall.Bind(fd, &syscall.SockaddrInet4{
			Port: laddr.Port,
			Addr: ipv4AddrToArray(laddr.IP),
		})
	} else {
		err = syscall.Bind(fd, &syscall.SockaddrInet6{
			Port: laddr.Port,
			Addr: ipv6AddrToArray(laddr.IP),
		})
	}
	if err != nil {
		syscall.Close(fd)
		return nil, errors.Wrap(err, "bind failed")
	}

	err = syscall.Listen(fd, 128)
	if err != nil {
		syscall.Close(fd)
		return nil, errors.Wrap(err, "listen failed")
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
	fd, sa, err := syscall.Accept(l.fd)
	if err != nil {
		return nil, err
	}

	raddr := &net.TCPAddr{
		Port: 0,
	}

	switch sa.(type) {
	case *syscall.SockaddrInet4:
		x := sa.(*syscall.SockaddrInet4)
		raddr.IP = net.IP(x.Addr[:])
		raddr.Port = x.Port
	case *syscall.SockaddrInet6:
		x := sa.(*syscall.SockaddrInet4)
		raddr.IP = net.IP(x.Addr[:])
		raddr.Port = x.Port
	}

	return &Conn{
		fd:    fd,
		laddr: l.laddr,
		raddr: raddr,
	}, nil
}
