package config

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/bio-routing/tflow2/convert"
)

type Global struct {
	LocalASN         uint32
	RouterID         uint32
	Port             uint16
	LocalAddressList []net.IP
	Listen           bool
}

const BGPPORT = uint16(179)

func (g *Global) SetDefaultGlobalConfigValues() error {
	if g.LocalAddressList == nil {
		g.LocalAddressList = make([]net.IP, 0)
		g.LocalAddressList = append(g.LocalAddressList, net.ParseIP("0.0.0.0"))
		g.LocalAddressList = append(g.LocalAddressList, net.ParseIP("::"))
	}

	if g.RouterID == 0 {
		rtrid, err := generateRouterID()
		if err != nil {
			return fmt.Errorf("unable to determine router ID: %w", err)
		}
		g.RouterID = rtrid
	}

	if g.Port == 0 {
		g.Port = BGPPORT
	}

	return nil
}

func generateRouterID() (uint32, error) {
	addr, err := getLoopbackIP()
	if err == nil {
		return convert.Uint32b([]byte(addr)[12:16]), nil
	}

	return 0, fmt.Errorf("unable to determine router id")
}

func getHighestIP() (net.IP, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("unable to ")
	}

	return _getHighestIP(ifs)
}

func _getHighestIP(ifs []net.Interface) (net.IP, error) {
	candidates := make([]net.IP, 0)
	for _, iface := range ifs {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("unable to get interface addrs for %s: %w", iface.Name, err)
		}

		for _, addr := range addrs {
			a := net.ParseIP(addr.String())
			if addr.String() != "127.0.0.1" && a.To4() != nil {
				candidates = append(candidates, a)
			}
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("No IPv4 address found on any interface")
	}

	max := candidates[0]
	for _, c := range candidates[1:] {
		if addrIsGreater(c, max) {
			max = c
		}
	}

	return nil, fmt.Errorf("No non localhost IPv4 address found on interface lo")
}

func getLoopbackIP() (net.IP, error) {
	iface, err := net.InterfaceByName("lo")
	if err != nil {
		return nil, fmt.Errorf("unable to get interface lo: %w", err)
	}

	return _getLoopbackIP(iface)
}

func _getLoopbackIP(iface *net.Interface) (net.IP, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("unable to get interface addresses: %w", err)
	}

	candidates := make([]net.IP, 0)
	for _, addr := range addrs {
		a := net.ParseIP(strings.Split(addr.String(), "/")[0])
		if a.String() != "127.0.0.1" && a.To4() != nil {
			candidates = append(candidates, a)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("No non localhost IPv4 address found on interface lo")
	}

	max := candidates[0]
	for _, c := range candidates {
		if addrIsGreater(c, max) {
			max = c
		}
	}

	return max, nil
}

// addrIsGreater returns true if the first address is higher than the second.
// Returns false if not.
func addrIsGreater(a net.IP, b net.IP) bool {
	if bytes.Compare(a, b) == 1 {
		return true
	}
	return false
}
