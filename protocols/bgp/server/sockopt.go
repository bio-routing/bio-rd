package server

import (
	"net"

	"github.com/bio-routing/bio-rd/net/tcp"
)

func setTTL(c net.Conn, ttl uint8) error {
	// as c is an interface for testability reason we're checking here if the concrete type
	// is a real TCP connection as only that supports setting a TTL
	switch c.(type) {
	case *tcp.Conn:
		return c.(*tcp.Conn).SetTTL(ttl)
	default:
		return nil
	}
}

func setDontRoute(c net.Conn) error {
	// as c is an interface for testability reason we're checking here if the concrete type
	// is a real TCP connection as only that supports setting SetDontRoute()
	switch c.(type) {
	case *tcp.Conn:
		return c.(*tcp.Conn).SetDontRoute()
	default:
		return nil
	}
}
