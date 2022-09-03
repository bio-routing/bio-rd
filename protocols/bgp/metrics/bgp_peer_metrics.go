package metrics

import (
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
)

const (
	StateDown        = 0
	StateIdle        = 1
	StateConnect     = 2
	StateActive      = 3
	StateOpenSent    = 4
	StateOpenConfirm = 5
	StateEstablished = 6
)

// BGPPeerMetrics provides metrics for one BGP session
type BGPPeerMetrics struct {
	// IP is the remote IP of the peer
	IP *bnet.IP

	// ASN is the ASN of the peer
	ASN uint32

	// LocalASN is our local ASN
	LocalASN uint32

	// VRF is the name of the VRF the peer is configured in
	VRF string

	// Since is the time the session was established
	Since time.Time

	// State of the BGP session (Down = 0, Idle = 1, Connect = 2, Active = 3, OpenSent = 4, OpenConfirm = 5, Established = 6)
	State uint8

	// UpdatesReceived is the number of update messages received on this session
	UpdatesReceived uint64

	// UpdatesReceived is the number of update messages we sent on this session
	UpdatesSent uint64

	// AddressFamilies provides metrics on AFI/SAFI level
	AddressFamilies []*BGPAddressFamilyMetrics
}
