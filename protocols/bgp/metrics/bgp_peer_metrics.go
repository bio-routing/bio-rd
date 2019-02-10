package metrics

import (
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
)

type BGPPeerMetrics struct {
	IP              bnet.IP
	Since           time.Duration
	ASN             uint32
	LocalASN        uint32
	UpdatesReceived uint64
	UpdatesSent     uint64
	AddressFamilies []*BGPAddressFamilyMetrics
}
