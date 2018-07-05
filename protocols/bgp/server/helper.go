package server

import (
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
)

func afiForPrefix(pfx bnet.Prefix) uint16 {
	if pfx.Addr().IsIPv4() {
		return packet.IPv6AFI
	}

	return packet.IPv6AFI
}
