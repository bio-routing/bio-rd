package net

import "net"

// IPv4ToUint32 converts an `net.IP` to an uint32 interpretation
func IPv4ToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[3]) + uint32(ip[2])<<8 + uint32(ip[1])<<16 + uint32(ip[0])<<24
}
