package server

import (
	"net"
	"time"
)

type fakeConn struct {
}

type fakeAddr struct {
}

func (f fakeAddr) Network() string {
	return ""
}

func (f fakeAddr) String() string {
	return "169.254.100.100:179"
}

func (f fakeConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (f fakeConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (f fakeConn) Close() error {
	return nil
}

func (f fakeConn) LocalAddr() net.Addr {
	return fakeAddr{}
}

func (f fakeConn) RemoteAddr() net.Addr {
	return fakeAddr{}
}

func (f fakeConn) SetDeadline(t time.Time) error {
	return nil
}

func (f fakeConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (f fakeConn) SetWriteDeadline(t time.Time) error {
	return nil
}
