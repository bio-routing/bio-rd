package netflow

import "net"

// ToIPNet returns the net.IPNet representation for the Prefix
func (pfx *Pfx) ToIPNet() *net.IPNet {
	return &net.IPNet{
		IP:   pfx.IP,
		Mask: pfx.Mask,
	}
}
