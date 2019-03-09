package route

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
)

const (
	// ProtoUnspec equals to unspec from /etc/iproute2/rt_protos
	ProtoUnspec = 0

	// ProtoRedirect equals to redirect from /etc/iproute2/rt_protos
	ProtoRedirect = 1

	// ProtoKernel equals to kernel from /etc/iproute2/rt_protos
	ProtoKernel = 2

	// ProtoBoot equals to boot from /etc/iproute2/rt_protos
	ProtoBoot = 3

	// ProtoStatic equals to static from /etc/iproute2/rt_protos
	ProtoStatic = 4

	// ProtoZebra equals to zebra from /etc/iproute2/rt_protos
	ProtoZebra = 11

	// ProtoBird equals to bird from /etc/iproute2/rt_protos
	ProtoBird = 12

	// ProtoDHCP equals to dhcp from /etc/iproute2/rt_protos
	ProtoDHCP = 16

	// ProtoBio bio-rd
	ProtoBio = 45
)

// FIBPath represents a path learned via Netlink of a route
type FIBPath struct {
	Src      bnet.IP
	NextHop  bnet.IP // GW
	Priority int
	Protocol int
	Type     int
	Table    int
	Kernel   bool // True if the route is already installed in the kernel
}

// NewNlPathFromBgpPath creates a new FIBPath object from a BGPPath object
func NewNlPathFromBgpPath(p *BGPPath) *FIBPath {
	return &FIBPath{
		Src:      p.Source,
		NextHop:  p.NextHop,
		Protocol: ProtoBio,
		Kernel:   false,
	}
}

// Select compares s with t and returns negative if s < t, 0 if paths are equal, positive if s > t
func (s *FIBPath) Select(t *FIBPath) int8 {
	if s.NextHop.Compare(t.NextHop) < 0 {
		return -1
	}

	if s.NextHop.Compare(t.NextHop) > 0 {
		return 1
	}

	if s.Src.Compare(t.Src) < 0 {
		return -1
	}

	if s.Src.Compare(t.Src) > 0 {
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
func (s *FIBPath) ECMP(t *FIBPath) bool {
	return s.Src == t.Src && s.Priority == t.Priority && s.Protocol == t.Protocol && s.Type == t.Type && s.Table == t.Table
}

// Copy duplicates the current object
func (s *FIBPath) Copy() *FIBPath {
	if s == nil {
		return nil
	}

	cp := *s
	return &cp
}

// Print all known information about a route in logfile friendly format
func (s *FIBPath) String() string {
	ret := fmt.Sprintf("Source: %s, ", s.Src.String())
	ret += fmt.Sprintf("NextHop: %s, ", s.NextHop.String())
	ret += fmt.Sprintf("Priority: %d, ", s.Priority)
	ret += fmt.Sprintf("Type: %d, ", s.Type)
	ret += fmt.Sprintf("Table: %d", s.Table)

	return ret
}

// Print all known information about a route in human readable form
func (s *FIBPath) Print() string {
	ret := fmt.Sprintf("\t\tSource: %s\n", s.Src.String())
	ret += fmt.Sprintf("\t\tNextHop: %s\n", s.NextHop.String())
	ret += fmt.Sprintf("\t\tPriority: %d\n", s.Priority)
	ret += fmt.Sprintf("\t\tType: %d\n", s.Type)
	ret += fmt.Sprintf("\t\tTable: %d\n", s.Table)

	return ret
}
