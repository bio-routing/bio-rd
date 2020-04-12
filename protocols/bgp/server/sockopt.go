package server

import (
	"net"

	"github.com/bio-routing/bio-rd/net/tcp"
)

func setTTL(c net.Conn, ttl uint8) error {
	switch c.(type) {
	case *tcp.Conn:
	default:
		return nil
	}

	return c.(*tcp.Conn).SetTTL(ttl)
}

func setDontRoute(c net.Conn) error {
	switch c.(type) {
	case *tcp.Conn:
	default:
		return nil
	}

	return c.(*tcp.Conn).SetDontRoute()
}
