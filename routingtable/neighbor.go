package routingtable

import bnet "github.com/bio-routing/bio-rd/net"

// Neighbor represents the attributes identifying a neighbor relationsship
type Neighbor struct {
	// Addres is the IPv4 address of the neighbor as integer representation
	Address bnet.IP

	// Local address is the local address of the BGP TCP connection
	LocalAddress bnet.IP

	// Type is the type / protocol used for routing inforation communitation
	Type uint8

	// IBGP returns if local ASN is equal to remote ASN
	IBGP bool

	// Local ASN of session
	LocalASN uint32

	// Peer is a route server client
	RouteServerClient bool

	// Peer is a route reflector client
	RouteReflectorClient bool

	// Our route reflection clusterID
	ClusterID uint32

	// CapAddPathRX indicates if the peer supports receiving multiple BGP paths
	CapAddPathRX bool
}
