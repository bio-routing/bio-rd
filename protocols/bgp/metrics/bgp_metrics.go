package metrics

import bnet "github.com/bio-routing/bio-rd/net"

type BGPMetrics struct {
	RouterID     bnet.IP
	LocalASN     uint32
	OpenReceived uint64
	OpenSent     uint64
	Neighbors    []*BGPMetrics
}
