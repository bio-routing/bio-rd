package net

import (
	"net"

	bnet "github.com/bio-routing/bio-rd/net"
)

// BIONetIPFromAddr retrives the IP from an net.Addr and returns the IP in BIO's internal IP type
func BIONetIPFromAddr(hostPort string) (bnet.IP, error) {
	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return bnet.IP{}, err
	}

	return bnet.IPFromString(host)
}
