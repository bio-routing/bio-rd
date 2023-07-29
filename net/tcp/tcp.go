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

type ConnI interface {
	Write(b []byte) (n int, err error)
	Read(b []byte) (n int, err error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	SetTTL(ttl uint8) error
	SetDontRoute() error
	SetNoDelay() error
	SetBindToDev(devName string) error
}

// Conn is TCP connection
type Conn struct {
	fd    int
	laddr *net.TCPAddr
	raddr *net.TCPAddr
}

// Dial established a new TCP connection
func Dial(laddr, raddr *net.TCPAddr, ttl uint8, md5Secret string, noRoute bool, bindDev string) (*Conn, error) {
	if raddr == nil {
		return nil, fmt.Errorf("raddr is mandatory")
	}

	afi := uint16(unix.AF_INET)
	if raddr.IP.To4() == nil {
		afi = unix.AF_INET6
	}

	c, err := dialTCP(afi, laddr, raddr, ttl, md5Secret, noRoute, bindDev)
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

// SetBindToDev sets the SO_BINDTODEVICE option
func (c *Conn) SetBindToDev(devName string) error {
	return bindToDev(c.fd, devName)
}

// MockConn is mocked TCP connection
type MockConn struct {
	chOut  chan []byte
	chIn   chan byte
	laddr  *net.TCPAddr
	raddr  *net.TCPAddr
	closed bool
}

func NewMockConn(srcIP net.IP, srcPort uint16, dstIP net.IP, dstPort uint16) *MockConn {
	return &MockConn{
		chOut: make(chan []byte, 10),
		chIn:  make(chan byte, 1000),
		laddr: &net.TCPAddr{
			IP:   srcIP,
			Port: int(srcPort),
		},
		raddr: &net.TCPAddr{
			IP:   dstIP,
			Port: int(dstPort),
		},
	}
}

// Write writes to a TCP connection
func (c *MockConn) Write(b []byte) (n int, err error) {
	if c.closed {
		return 0, fmt.Errorf("connection is closed")
	}

	c.chOut <- b
	return len(b), nil
}

// Read reads from a TCP connection
func (c *MockConn) Read(b []byte) (n int, err error) {
	if c.closed {
		return 0, fmt.Errorf("connection is closed")
	}

	for i := range b {
		b[i] = <-c.chIn
	}

	return len(b), nil
}

func (c *MockConn) WriteFromOtherEnd(b []byte) {
	for _, x := range b {
		c.chIn <- x
	}
}

func (c *MockConn) ReadFromOtherEnd() []byte {
	return <-c.chOut
}

func (c *MockConn) Close() error {
	c.closed = true
	return nil
}

func (c *MockConn) LocalAddr() net.Addr {
	return c.laddr
}

func (c *MockConn) RemoteAddr() net.Addr {
	return c.raddr
}

// SetDeadline is here to fulfill net.Conn interface
func (c *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (c *MockConn) SetTTL(ttl uint8) error {
	return nil
}

func (c *MockConn) SetDontRoute() error {
	return nil
}

func (c *MockConn) SetNoDelay() error {
	return nil
}

func (c *MockConn) SetBindToDev(devName string) error {
	return nil
}
