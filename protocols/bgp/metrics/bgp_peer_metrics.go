package metrics

import (
	bnet "github.com/bio-routing/bio-rd/net"
	"time"
)

// BGPPeerMetrics provides metrics for one BGP session
type BGPPeerMetrics struct {
	// IP is the remote IP of the peer
	IP bnet.IP

	// Since is the time the session was established
	Since time.Time

	// Status of the BGP session (Down = 0, Idle = 1, Connect = 2, OpenSent = 3, OpenConfirm = 4, Established = 5)
	Status uint8

	// Up returns if the session is established
	Up bool

	// ASN is the ASN of the peer
	ASN uint32

	// LocalASN is our local ASN
	LocalASN uint32

	// VRF is the name of the VRF the peer is configured in
	VRF string

	// UpdatesReceived is the number of update messages received on this session
	UpdatesReceived uint64

	// UpdatesReceived is the number of update messages we sent on this session
	UpdatesSent uint64

	// AddressFamilies provides metrics on AFI/SAFI level
	AddressFamilies []*BGPAddressFamilyMetrics
}
