package server

import (
	"net"

	"github.com/bio-routing/bio-rd/net/tcp"
	"github.com/bio-routing/bio-rd/util/log"
)

const (
	// BGPPORT is the port of the BGP protocol
	BGPPORT = 179
)

// TCPListener is a TCP listen wrapper
type TCPListener struct {
	l       *tcp.Listener
	addr    *net.TCPAddr
	closeCh chan struct{}
}

// NewTCPListener creates a new TCPListener
func NewTCPListener(addr string) (*TCPListener, error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	l, err := tcp.Listen(tcpaddr, 255)
	if err != nil {
		return nil, err
	}

	tl := &TCPListener{
		l:       l,
		addr:    tcpaddr,
		closeCh: make(chan struct{}),
	}

	return tl, nil
}

func (t *TCPListener) Accept() (net.Conn, error) {
	conn, err := t.l.AcceptTCP()
	if err != nil {
		close(t.closeCh)

		log.WithError(err).WithFields(log.Fields{
			"Topic": "Peer",
		}).Error("Failed to AcceptTCP")

		return nil, err
	}

	return conn, nil
}

func (t *TCPListener) Addr() net.Addr {
	return t.addr
}

func (t *TCPListener) Close() error {
	close(t.closeCh)

	return nil
}

func (t *TCPListener) setTCPMD5(addr net.IP, secret string) error {
	return t.l.SetTCPMD5(addr, secret)
}
