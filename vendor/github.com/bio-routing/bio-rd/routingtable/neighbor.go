package routingtable

import bnet "github.com/bio-routing/bio-rd/net"

// Neighbor represents the attributes identifying a neighbor relationship
type Neighbor struct {
	// Address is the IPv4 address of the neighbor as integer representation
	Address *bnet.IP

	// Local address is the local address of the BGP TCP connection
	LocalAddress *bnet.IP

	// Type is the type / protocol used for routing inforation communitation
	Type uint8

	// IBGP returns if local ASN is equal to remote ASN
	IBGP bool

	// Local ASN of session
	LocalASN uint32

	// RouteServerClient indicates if the peer is a route server client
	RouteServerClient bool

	// RouteReflectorClient indicates if the peer is a route reflector client
	RouteReflectorClient bool

	// ClusterID is our route reflectors clusterID
	ClusterID uint32
}
