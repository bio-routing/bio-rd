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

func (m *MockHandler) NewConn(dest [EthALen]byte) net.Conn {
	return &MockConn{
		eth:      m,
		destAddr: dest,
		C:        btesting.NewMockConnBidi(&btesting.MockAddr{}, &btesting.MockAddr{}),
	}
}
