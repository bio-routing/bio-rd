package route

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// ProtoBio is the protocol number constant for bio from /etc/iproute2/rt_protos
const ProtoBio = 45

// NetlinkPath represents a path learned via Netlink of a route
type NetlinkPath struct {
	Dst      bnet.Prefix
	Src      bnet.IP
	NextHop  bnet.IP // GW
	Priority int
	Protocol int
	Type     int
	Table    int
	Kernel   bool // True if the route is already installed in the kernel
}

// NewNlPathFromBgpPath creates a new NetlinkPath object from a BGPPath object
func NewNlPathFromBgpPath(p *BGPPath) *NetlinkPath {
	return &NetlinkPath{
		Src:      p.Source,
		NextHop:  p.NextHop,
		Protocol: ProtoBio,
		Kernel:   false,
	}
}

// NewNlPathFromRoute creates a new NetlinkPath object from a netlink.Route object
func NewNlPathFromRoute(r *netlink.Route, kernel bool) (*NetlinkPath, error) {
	var src bnet.IP
	var dst bnet.Prefix

	if r.Src == nil && r.Dst == nil {
		return nil, fmt.Errorf("Cannot create NlPath, since source and destination are both nil")
	}

	if r.Src == nil && r.Dst != nil {
		dst = bnet.NewPfxFromIPNet(r.Dst)
		if dst.Addr().IsIPv4() {
			src = bnet.IPv4FromOctets(0, 0, 0, 0)
		} else {
			src = bnet.IPv6FromBlocks(0, 0, 0, 0, 0, 0, 0, 0)
		}
	}

	if r.Src != nil && r.Dst == nil {
		src, _ = bnet.IPFromBytes(r.Src)
		if src.IsIPv4() {
			dst = bnet.NewPfx(bnet.IPv4FromOctets(0, 0, 0, 0), 0)
		} else {
			dst = bnet.NewPfx(bnet.IPv6FromBlocks(0, 0, 0, 0, 0, 0, 0, 0), 0)
		}
	}

	if r.Src != nil && r.Dst != nil {
		src, _ = bnet.IPFromBytes(r.Src)
		dst = bnet.NewPfxFromIPNet(r.Dst)
	}

	log.Warnf("IPFromBytes: %v goes to %v", r.Src, src)
	log.Warnf("IPFromBytes: %v goes to %v", r.Dst, dst)

	nextHop, _ := bnet.IPFromBytes(r.Gw)

	return &NetlinkPath{
		Dst:      dst,
		Src:      src,
		NextHop:  nextHop,
		Priority: r.Priority,
		Protocol: r.Protocol,
		Type:     r.Type,
		Table:    r.Table,
		Kernel:   kernel,
	}, nil
}

// Select compares s with t and returns negative if s < t, 0 if paths are equal, positive if s > t
func (s *NetlinkPath) Select(t *NetlinkPath) int8 {
	if !s.Dst.Equal(t.Dst) {
		return 1
	}

	if s.NextHop.Compare(t.NextHop) > 0 {
		return -1
	}

	if s.NextHop.Compare(t.NextHop) < 0 {
		return 1
	}

	if s.Src.Compare(t.Src) > 0 {
		return -1
	}

	if s.Src.Compare(t.Src) < 0 {
		return 1
	}

	if s.Priority < t.Priority {
		return -1
	}

	if s.Priority > t.Priority {
		return 1
	}

	if s.Protocol < t.Protocol {
		return -1
	}

	if s.Protocol > t.Protocol {
		return 1
	}

	if s.Table < t.Table {
		return -1
	}

	if s.Table > t.Table {
		return 1
	}

	return 0
}

// ECMP determines if path s and t are equal in terms of ECMP
func (s *NetlinkPath) ECMP(t *NetlinkPath) bool {
	return false
}

// Copy duplicates the current object
func (s *NetlinkPath) Copy() *NetlinkPath {
	if s == nil {
		return nil
	}

	cp := *s
	return &cp
}

// Print all known information about a route in logfile friendly format
func (s *NetlinkPath) String() string {
	ret := fmt.Sprintf("Destination: %s, ", s.Dst.String())
	ret += fmt.Sprintf("Source: %s, ", s.Src.String())
	ret += fmt.Sprintf("NextHop: %s, ", s.NextHop.String())
	ret += fmt.Sprintf("Priority: %d, ", s.Priority)
	ret += fmt.Sprintf("Type: %d, ", s.Type)
	ret += fmt.Sprintf("Table: %d", s.Table)

	return ret
}

// Print all known information about a route in human readable form
func (s *NetlinkPath) Print() string {
	ret := fmt.Sprintf("\t\tDestination: %s\n", s.Dst.String())
	ret += fmt.Sprintf("\t\tSource: %s\n", s.Src.String())
	ret += fmt.Sprintf("\t\tNextHop: %s\n", s.NextHop.String())
	ret += fmt.Sprintf("\t\tPriority: %d\n", s.Priority)
	ret += fmt.Sprintf("\t\tType: %d\n", s.Type)
	ret += fmt.Sprintf("\t\tTable: %d\n", s.Table)

	return ret
}
