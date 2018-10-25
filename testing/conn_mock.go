package testing

import (
	"bytes"
	"net"
)

// MockConn mock an connection
type MockConn struct {
	net.Conn
	Buf    *bytes.Buffer
	Closed bool
}

func NewMockConn() *MockConn {
	return &MockConn{
		Buf: bytes.NewBuffer(nil),
	}
}

func (m *MockConn) Write(b []byte) (int, error) {
	return m.Buf.Write(b)
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	return m.Buf.Read(b)
}

func (m *MockConn) Close() error {
	m.Closed = true
	return nil
}
