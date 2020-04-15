package tcp

import (
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// Conn is TCP connection
type Conn struct {
	fd    int
	laddr *net.TCPAddr
	raddr *net.TCPAddr
}

// Dial established a new TCP connection
func Dial(laddr, raddr *net.TCPAddr, ttl uint8, md5Secret string, noRoute bool) (*Conn, error) {
	if raddr == nil {
		return nil, fmt.Errorf("raddr is mandatory")
	}

	afi := uint16(syscall.AF_INET)
	if raddr.IP.To4() == nil {
		afi = syscall.AF_INET6
	}

	c, err := dialTCP(afi, laddr, raddr, ttl, md5Secret, noRoute)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing failed")
	}

	c.laddr = laddr
	if c.laddr == nil || c.laddr.IP == nil {
		sa, err := syscall.Getsockname(c.fd)
		if err != nil {
			return nil, errors.Wrap(err, "getsockname() failed")
		}

		sa4 := sa.(*syscall.SockaddrInet4)
		c.laddr.IP = net.IP(sa4.Addr[:])
		c.laddr.Port = sa4.Port
	}
	c.raddr = raddr
	return c, nil
}

// Write writes to a TCP connection
func (c *Conn) Write(b []byte) (n int, err error) {
	return syscall.Write(c.fd, b)
}

// Read reads from a TCP connection
func (c *Conn) Read(b []byte) (n int, err error) {
	return syscall.Read(c.fd, b)
}

// Close closes the connection
func (c *Conn) Close() error {
	return syscall.Close(c.fd)
}

// LocalAddr gets the local address
func (c *Conn) LocalAddr() net.Addr {
	return c.laddr
}

// RemoteAddr gets the remote address
func (c *Conn) RemoteAddr() net.Addr {
	return c.raddr
}

// SetDeadline is here to fulfill net.Conn interface
func (c *Conn) SetDeadline(t time.Time) error {
	return fmt.Errorf("No supported")
}

// SetReadDeadline is here to fulfill net.Conn interface
func (c *Conn) SetReadDeadline(t time.Time) error {
	return fmt.Errorf("No supported")
}

// SetWriteDeadline is here to fulfill net.Conn interface
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return fmt.Errorf("No supported")
}

// SetTTL sets the TTL on a TCP connection
func (c *Conn) SetTTL(ttl uint8) error {
	if c.raddr.IP.To4() != nil {
		return syscall.SetsockoptInt(c.fd, syscall.IPPROTO_IP, syscall.IP_TTL, int(ttl))
	}

	return syscall.SetsockoptInt(c.fd, syscall.IPPROTO_IPV6, syscall.IPV6_UNICAST_HOPS, int(ttl))
}

// SetDontRoute sets the SO_DONTROUTE option
func (c *Conn) SetDontRoute() error {
	return syscall.SetsockoptInt(c.fd, syscall.SOL_SOCKET, syscall.SO_DONTROUTE, 1)
}

// SetNoDelay sets the TCP_NODELAY option
func (c *Conn) SetNoDelay() error {
	return syscall.SetsockoptInt(c.fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
}
