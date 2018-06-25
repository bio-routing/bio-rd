package routingtable

// Neighbor represents the attributes identifying a neighbor relationsship
type Neighbor struct {
	// Addres is the IPv4 address of the neighbor as integer representation
	Address uint32

	// Local address is the local address of the BGP TCP connection
	LocalAddress uint32

	// Type is the type / protocol used for routing inforation communitation
	Type uint8

	// IBGP returns if local ASN is equal to remote ASN
	IBGP bool

	// Local ASN of session
	LocalASN uint32

	// Peer is a route server client
	RouteServerClient bool

	// CapAddPathRX indicates if the peer supports receiving multiple BGP paths
	CapAddPathRX bool
}
