package metrics

import (
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
)

// BGPPeerMetrics provides metrics for one BGP session
type BGPPeerMetrics struct {
	// IP is the remote IP of the peer
	IP bnet.IP
	// Since is the duration the session is established
	Since time.Duration
	// ASN is the ASN of the peer
	ASN uint32
	// LocalASN is our local ASN
	LocalASN uint32
	// UpdatesReceived is the number of update messages received on this session
	UpdatesReceived uint64
	// UpdatesReceived is the number of update messages we sent on this session
	UpdatesSent uint64
	// AddressFamilies provides metrics on AFI/SAFI level
	AddressFamilies []*BGPAddressFamilyMetrics
}
