package testing

import (
	"net"
)

// MockConn mock an connection
type MockConn struct {
	net.Conn

	// Bytes are the bytes writen
	Bytes []byte
}

func NewMockConn() *MockConn {
	return &MockConn{
		Bytes: make([]byte, 0),
	}
}

func (m *MockConn) Write(b []byte) (int, error) {
	m.Bytes = append(m.Bytes, b...)
	return len(b), nil
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	count := len(b)
	if count > len(m.Bytes) {
		count = len(m.Bytes)
	}

	copy(b, m.Bytes[0:count])
	return count, nil
}
