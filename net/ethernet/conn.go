package ethernet

import (
	"fmt"
	"net"
	"time"
)

// EthAddr implements net.Addr
type EthAddr struct {
	address string
}

// String returns the address
func (ea EthAddr) String() string {
	return ea.address
}

// Network returns "ethernet"
func (ea EthAddr) Network() string {
	return "ethernet"
}

// Conn is a wrapper that implements net.Conn and sents packets to a particular destination
type Conn struct {
	eth      *Handler
	destAddr [EthALen]byte
}

// NewConn creates a new ethernet conn
func (e *Handler) NewConn(dest [EthALen]byte) net.Conn {
	return &Conn{
		eth:      e,
		destAddr: dest,
	}
}

// Read is there to fulfill the net.Conn interface
func (c *Conn) Read(b []byte) (n int, err error) {
	return 0, fmt.Errorf("Not supported")
}

// Write sends b on the Conn
func (c *Conn) Write(b []byte) (n int, err error) {
	err = c.eth.sendPacket(b, c.destAddr)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

// Close is here to fulfill the net.Conn interface
func (c *Conn) Close() error {
	return fmt.Errorf("Not supported")
}

// LocalAddr returns the local address
func (c *Conn) LocalAddr() net.Addr {
	return EthAddr{}
}

// RemoteAddr returns the remote address
func (c *Conn) RemoteAddr() net.Addr {
	return EthAddr{}
}

// SetDeadline is here to fulfill the net.Conn interface
func (c *Conn) SetDeadline(t time.Time) error {
	return fmt.Errorf("Not supported")
}

// SetReadDeadline is here to fulfill the net.Conn interface
func (c *Conn) SetReadDeadline(t time.Time) error {
	return fmt.Errorf("Not supported")
}

// SetWriteDeadline is here to fulfill the net.Conn interface
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return fmt.Errorf("Not supported")
}
