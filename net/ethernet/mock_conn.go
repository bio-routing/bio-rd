package ethernet

import (
	"fmt"
	"net"
	"time"

	btesting "github.com/bio-routing/bio-rd/testing"
)

// MockConn is a mock of Conn
type MockConn struct {
	eth      *MockHandler
	destAddr [ethALen]byte
	C        *btesting.MockConnBidi
}

// Read is there to fulfill the net.Conn interface
func (c *MockConn) Read(b []byte) (n int, err error) {
	return 0, fmt.Errorf("not supported")
}

// Write sends b on the Conn
func (c *MockConn) Write(b []byte) (n int, err error) {
	return c.C.Write(b)
}

// Close is here to fulfill the net.Conn interface
func (c *MockConn) Close() error {
	return fmt.Errorf("not supported")
}

// LocalAddr returns the local address
func (c *MockConn) LocalAddr() net.Addr {
	return EthAddr{}
}

// RemoteAddr returns the remote address
func (c *MockConn) RemoteAddr() net.Addr {
	return EthAddr{}
}

// SetDeadline is here to fulfill the net.Conn interface
func (c *MockConn) SetDeadline(t time.Time) error {
	return fmt.Errorf("not supported")
}

// SetReadDeadline is here to fulfill the net.Conn interface
func (c *MockConn) SetReadDeadline(t time.Time) error {
	return fmt.Errorf("not supported")
}

// SetWriteDeadline is here to fulfill the net.Conn interface
func (c *MockConn) SetWriteDeadline(t time.Time) error {
	return fmt.Errorf("not supported")
}
