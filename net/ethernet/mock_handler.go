package ethernet

import (
	"net"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	btesting "github.com/bio-routing/bio-rd/testing"
)

type MockHandler struct {
}

func NewMockHandler() HandlerInterface {
	return &MockHandler{}
}

func (m *MockHandler) NewConn(dest [EthALen]byte) net.Conn {
	return &MockConn{
		eth:      m,
		destAddr: dest,
		C:        btesting.NewMockConnBidi(&btesting.MockAddr{}, &btesting.MockAddr{}),
	}
}

// RecvPacket to be implemented
func (m *MockHandler) RecvPacket() (pkt []byte, src types.MACAddress, err error) {
	return nil, types.MACAddress{}, nil
}

func (m *MockHandler) MCastJoin(addr MACAddr) error {
	return nil
}
