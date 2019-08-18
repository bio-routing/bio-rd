package server

import (
	"net"

	log "github.com/sirupsen/logrus"
)

const (
	BGPPORT = 179
)

type TCPListener struct {
	l       *net.TCPListener
	closeCh chan struct{}
}

func NewTCPListener(addr string, ch chan net.Conn) (*TCPListener, error) {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	l, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return nil, err
	}

	// Note: Set TTL=255 for incoming connection listener in order to accept
	// connection in case for the neighbor has TTL Security settings.
	if err := SetListenTCPTTLSockopt(l, 255); err != nil {
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   addr,
		}).Warnf("cannot set TTL(=%d) for TCPListener: %s", 255, err)
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
				log.WithFields(log.Fields{
					"Topic": "Peer",
					"Error": err,
				}).Warn("Failed to AcceptTCP")
				return err
			}
			ch <- conn
		}
	}(tl)

	return tl, nil
}
