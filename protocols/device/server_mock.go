package device

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

func (ms *MockServer) DeviceUpEvent(name string) {
	ms.C.DeviceUpdate(&Device{
		name:      name,
		operState: IfOperUp,
	})
}
