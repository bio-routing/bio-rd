package ethernet

import (
	"net"

	btesting "github.com/bio-routing/bio-rd/testing"
)

type MockHandler struct {
}

func NewMockHandler() HandlerInterface {
	return &MockHandler{}
}

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

func (m *MockHandler) MCastJoin(addr MACAddr) error {
	return nil
}

func (m *MockHandler) GetMTU() int {
	return 1500
}

func (m *MockHandler) SendPacket(dst MACAddr, pkt []byte) error {
	return nil
}
