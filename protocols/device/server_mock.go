package device

import "github.com/bio-routing/bio-rd/net"

type MockServer struct {
	Called            bool
	UnsubscribeCalled bool
	C                 Client
	Name              string
	UnsubscribeName   string
}

func (ms *MockServer) Start() error {
	return nil
}

func (ms *MockServer) Subscribe(c Client, n string) {
	ms.C = c
	ms.Called = true
	ms.Name = n
}

func (ms *MockServer) Unsubscribe(c Client, n string) {
	ms.UnsubscribeCalled = true
	ms.UnsubscribeName = n
}

func (ms *MockServer) DeviceUpEvent(name string, addrs []*net.Prefix) {
	ms.C.DeviceUpdate(&Device{
		name:      name,
		operState: IfOperUp,
		addrs:     addrs,
	})
}

func (ms *MockServer) DeviceDownEvent(name string, addrs []*net.Prefix) {
	ms.C.DeviceUpdate(&Device{
		name:      name,
		operState: IfOperDown,
		addrs:     addrs,
	})
}
