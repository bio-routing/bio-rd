package route

import (
	"fmt"
	"strings"

	bnet "github.com/bio-routing/bio-rd/net"
)

const (
	ProtoBio = 45 // bio
)

// FIBPath represents a path learned via Netlink of a route
type FIBPath struct {
	Src      *bnet.IP
	NextHop  *bnet.IP // GW
	Priority int
	Protocol int
	Type     int
	Table    int
	Kernel   bool // True if the route is already installed in the kernel
}

// NewNlPathFromBgpPath creates a new FIBPath object from a BGPPath object
func NewNlPathFromBgpPath(p *BGPPath) *FIBPath {
	return &FIBPath{
		Src:      p.BGPPathA.Source,
		NextHop:  p.BGPPathA.NextHop,
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
	return s.Src.Compare(t.Src) == 0 && s.Priority == t.Priority && s.Protocol == t.Protocol && s.Type == t.Type && s.Table == t.Table
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
	var b strings.Builder

	fmt.Fprintf(&b, "Source: %s, ", s.Src.String())
	fmt.Fprintf(&b, "NextHop: %s, ", s.NextHop.String())
	fmt.Fprintf(&b, "Priority: %d, ", s.Priority)
	fmt.Fprintf(&b, "Type: %d, ", s.Type)
	fmt.Fprintf(&b, "Table: %d", s.Table)

	return b.String()
}

// Print all known information about a route in human readable form
func (s *FIBPath) Print() string {
	var b strings.Builder

	fmt.Fprintf(&b, "\t\tSource: %s\n", s.Src.String())
	fmt.Fprintf(&b, "\t\tNextHop: %s\n", s.NextHop.String())
	fmt.Fprintf(&b, "\t\tPriority: %d\n", s.Priority)
	fmt.Fprintf(&b, "\t\tType: %d\n", s.Type)
	fmt.Fprintf(&b, "\t\tTable: %d\n", s.Table)

	return b.String()
}
