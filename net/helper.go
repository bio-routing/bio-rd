package net

import "net"

// IPv4ToUint32 converts an `net.IP` to an uint32 interpretation
func IPv4ToUint32(ip net.IP) uint32 {
	return uint32(ip[0]) + uint32(ip[1])<<8 + uint32(ip[2])<<16 + uint32(ip[3])<<24
}
