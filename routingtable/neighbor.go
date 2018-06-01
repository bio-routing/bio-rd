package routingtable

import "fmt"

// Neighbor represents the attributes identifying a neighbor relationsship
type Neighbor struct {
	// Addres is the IPv4 address of the neighbor as integer representation
	Address uint32

	// Type is the type / protocol used for routing inforation communitation
	Type uint8
}

func (n *Neighbor) String() string {
	return fmt.Sprintf("%d-%d", n.Address, n.Type)
}
