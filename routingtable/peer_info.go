package routingtable

import bnet "github.com/bio-routing/bio-rd/net"

// PeerInfo represents the attributes identifying a neighbor relationship
type PeerInfo struct {
	// RouterID is the ID of the local router
	RouterID uint32

	// PeerIP is the IP address of the neighbor
	PeerIP *bnet.IP

	// LocalIP is the local address of the BGP TCP connection
	LocalIP *bnet.IP

	// Type is the type / protocol used for routing inforation communitation
	Type uint8

	// IBGP returns if local ASN is equal to remote ASN
	IBGP bool

	// Local ASN of session
	LocalASN uint32

	// Peer ASN for this neighbor
	PeerASN uint32

	// RouteServerClient indicates if the peer is a route server client
	RouteServerClient bool

	// RouteReflectorClient indicates if the peer is a route reflector client
	RouteReflectorClient bool

	// ClusterID is our route reflectors clusterID
	ClusterID uint32

	// AddPathRX indicates if AddPath receive is active
	AddPathRX bool

	// AddPathRX indicates if AddPath receive is active
	AddPathTX bool

	// RouterIP indicates the IP address of the remote BMP peer (only for BMP)
	RouterIP bnet.IP
}
