package testing

import (
	"bytes"
	"net"
	"sync"
)

// MockConn mock an connection
type MockConn struct {
	net.Conn
	buf    *bytes.Buffer
	lock   sync.Mutex
	closed bool
}

func NewMockConn() *MockConn {
	return &MockConn{
		buf: bytes.NewBuffer(nil),
	}
}

func NewMockConnClosed() *MockConn {
	m := NewMockConn()
	m.buf = bytes.NewBuffer(nil)
	m.closed = true
	return m
}

func (m *MockConn) Write(b []byte) (int, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.buf.Write(b)
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.buf.Read(b)
}

func (m *MockConn) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.buf = bytes.NewBuffer(nil)
	m.closed = true
	return nil
}

func (m *MockConn) Closed() bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.closed
}
