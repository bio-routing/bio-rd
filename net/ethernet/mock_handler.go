package ethernet

import (
	"net"

	btesting "github.com/bio-routing/bio-rd/testing"
)

// MockHandler mocks an ethernet handler
type MockHandler struct {
}

// NewMockHandler creates a new mock handler
func NewMockHandler() HandlerInterface {
	return &MockHandler{}
}

// NewConn creates a new mocked ethernet conn
func (m *MockHandler) NewConn(dest MACAddr) net.Conn {
	return &MockConn{
		eth:      m,
		destAddr: dest,
		C:        btesting.NewMockConnBidi(&btesting.MockAddr{}, &btesting.MockAddr{}),
	}
}

// RecvPacket to be implemented
func (m *MockHandler) RecvPacket() (pkt []byte, src MACAddr, err error) {
	return nil, MACAddr{}, nil
}

// MCastJoin is here to fulfill an interface
func (m *MockHandler) MCastJoin(addr MACAddr) error {
	return nil
}

// GetMTU is here to fulfill an interface
func (m *MockHandler) GetMTU() int {
	return 1500
}

// SendPacket is here to fulfill an interface
func (m *MockHandler) SendPacket(dst MACAddr, pkt []byte) error {
	return nil
}
