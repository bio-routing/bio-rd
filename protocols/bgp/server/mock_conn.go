package server

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type mockCon struct {
	closed     bool
	localAddr  net.Addr
	remoteAddr net.Addr
	buffer     bytes.Buffer
}

type mockAddr struct {
}

func (m *mockAddr) Network() string {
	return ""
}

func (m *mockAddr) String() string {
	return ""
}

func newMockCon(localAddr net.Addr, remoteAddr net.Addr) *mockCon {
	return &mockCon{}
}

func (m *mockCon) Read(b []byte) (n int, err error) {

	return 0, nil
}

func (m *mockCon) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (m *mockCon) Close() error {
	m.closed = true
	return nil
}

func (m *mockCon) LocalAddr() net.Addr {
	return m.localAddr
}

func (m *mockCon) RemoteAddr() net.Addr {
	return m.remoteAddr
}

func (m *mockCon) SetDeadline(t time.Time) error {
	return fmt.Errorf("Not implemented")
}

func (m *mockCon) SetReadDeadline(t time.Time) error {
	return fmt.Errorf("Not implemented")
}

func (m *mockCon) SetWriteDeadline(t time.Time) error {
	return fmt.Errorf("Not implemented")
}
