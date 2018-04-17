package server

import (
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	BGPPORT = 179
)

type TCPListener struct {
	l       *net.TCPListener
	closeCh chan struct{}
}

func NewTCPListener(address net.IP, port uint16, ch chan *net.TCPConn) (*TCPListener, error) {
	proto := "tcp4"
	if address.To4() == nil {
		proto = "tcp6"
	}

	addr, err := net.ResolveTCPAddr(proto, net.JoinHostPort(address.String(), strconv.Itoa(int(port))))
	if err != nil {
		return nil, err
	}

	l, err := net.ListenTCP(proto, addr)
	if err != nil {
		return nil, err
	}

	// Note: Set TTL=255 for incoming connection listener in order to accept
	// connection in case for the neighbor has TTL Security settings.
	if err := SetListenTCPTTLSockopt(l, 255); err != nil {
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   addr,
		}).Warnf("cannot set TTL(=%d) for TCPLisnter: %s", 255, err)
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
