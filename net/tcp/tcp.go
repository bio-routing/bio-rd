package tcp

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/sys/unix"
)

const (
	// SOL_IP is not defined on darwin
	SOL_IP = 0x0

	// SOL_IPV6 is not defined on darwin
	SOL_IPV6 = 0x29
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

	afi := uint16(unix.AF_INET)
	if raddr.IP.To4() == nil {
		afi = unix.AF_INET6
	}

	c, err := dialTCP(afi, laddr, raddr, ttl, md5Secret, noRoute)
	if err != nil {
		return nil, fmt.Errorf("dialing failed: %w", err)
	}

	c.laddr = laddr
	if c.laddr == nil || c.laddr.IP == nil {
		sa, err := unix.Getsockname(c.fd)
		if err != nil {
			return nil, fmt.Errorf("getsockname() failed: %w", err)
		}

		sa4 := sa.(*unix.SockaddrInet4)
		c.laddr.IP = sa4.Addr[:]
		c.laddr.Port = sa4.Port
	}
	c.raddr = raddr
	return c, nil
}

// Write writes to a TCP connection
func (c *Conn) Write(b []byte) (n int, err error) {
	return unix.Write(c.fd, b)
}

// Read reads from a TCP connection
func (c *Conn) Read(b []byte) (n int, err error) {
	return unix.Read(c.fd, b)
}

// Close closes the connection
func (c *Conn) Close() error {
	return unix.Close(c.fd)
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
	return fmt.Errorf("not supported")
}

// SetReadDeadline is here to fulfill net.Conn interface
func (c *Conn) SetReadDeadline(t time.Time) error {
	return fmt.Errorf("not supported")
}

// SetWriteDeadline is here to fulfill net.Conn interface
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return fmt.Errorf("not supported")
}

// SetTTL sets the TTL on a TCP connection
func (c *Conn) SetTTL(ttl uint8) error {
	if c.raddr.IP.To4() != nil {
		return unix.SetsockoptInt(c.fd, SOL_IP, unix.IP_TTL, int(ttl))
	}

	return unix.SetsockoptInt(c.fd, unix.IPPROTO_IPV6, unix.IPV6_UNICAST_HOPS, int(ttl))
}

// SetDontRoute sets the SO_DONTROUTE option
func (c *Conn) SetDontRoute() error {
	return unix.SetsockoptInt(c.fd, unix.SOL_SOCKET, unix.SO_DONTROUTE, 1)
}

// SetNoDelay sets the TCP_NODELAY option
func (c *Conn) SetNoDelay() error {
	return unix.SetsockoptInt(c.fd, unix.IPPROTO_TCP, unix.TCP_NODELAY, 1)
}
