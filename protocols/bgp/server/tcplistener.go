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
	closeCh chan struct{}
}

// NewTCPListener creates a new TCPListener
func NewTCPListener(addr string, ch chan net.Conn) (*TCPListener, error) {
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
		closeCh: make(chan struct{}),
	}

	go func(tl *TCPListener) error {
		for {
			conn, err := tl.l.AcceptTCP()
			if err != nil {
				close(tl.closeCh)
				log.WithError(err).WithFields(log.Fields{
					"Topic": "Peer",
				}).Error("Failed to AcceptTCP")
				return err
			}
			ch <- conn
		}
	}(tl)

	return tl, nil
}

func (t *TCPListener) setTCPMD5(addr net.IP, secret string) error {
	return t.l.SetTCPMD5(addr, secret)
}
