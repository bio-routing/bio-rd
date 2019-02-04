package testing

import (
	"bytes"
	"net"
	"sync"
)

// MockConnBidi mock an connection
type MockConnBidi struct {
	net.Conn
	BufA   *bytes.Buffer
	BufAMu sync.Mutex
	BufB   *bytes.Buffer
	BufBMu sync.Mutex
	AddrA  *MockAddr
	AddrB  *MockAddr
	Closed bool
}

type MockAddr struct {
	Addr  string
	Proto string
}

func (ma *MockAddr) Network() string {
	return ma.Proto
}

func (ma *MockAddr) String() string {
	return ma.Addr
}

func NewMockConnBidi(addrA, addrB *MockAddr) *MockConnBidi {
	return &MockConnBidi{
		BufA:  bytes.NewBuffer(nil),
		BufB:  bytes.NewBuffer(nil),
		AddrA: addrA,
		AddrB: addrB,
	}
}

func (m *MockConnBidi) RemoteAddr() net.Addr {
	return m.AddrB
}

func (m *MockConnBidi) LocalAddr() net.Addr {
	return m.AddrA
}

func (m *MockConnBidi) Write(b []byte) (int, error) {
	m.BufAMu.Lock()
	defer m.BufAMu.Unlock()

	return m.BufA.Write(b)
}

func (m *MockConnBidi) Read(b []byte) (n int, err error) {
	m.BufBMu.Lock()
	defer m.BufBMu.Unlock()

	return m.BufB.Read(b)
}

func (m *MockConnBidi) WriteB(b []byte) (int, error) {
	m.BufBMu.Lock()
	defer m.BufBMu.Unlock()

	return m.BufB.Write(b)
}

func (m *MockConnBidi) ReadA(b []byte) (n int, err error) {
	m.BufAMu.Lock()
	defer m.BufAMu.Unlock()

	return m.BufA.Read(b)
}

func (m *MockConnBidi) Close() error {
	m.Closed = true
	return nil
}
