package metrics

import (
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
)

type BGPNeighborMetrics struct {
	IP              bnet.IP
	Since           time.Duration
	ASN             uint32
	AddressFamilies []*AddressFamilyMetrics
}
