package route

import (
	"fmt"
	"strings"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route/api"
)

const (
	HiddenReasonNone = iota
	HiddenReasonNextHopUnreachable
	HiddenReasonFilteredByPolicy
	HiddenReasonASLoop
	HiddenReasonOurOriginatorID
	HiddenReasonClusterLoop
	HiddenReasonOTCMismatch
)

// Path represents a network path
type Path struct {
	Type         uint8
	HiddenReason uint8  // If set, Path is hidden and ineligible to be installed in LocRIB and used for path selection
	LTime        uint32 // The time we learned this path, as unix epoch (seconds)
	StaticPath   *StaticPath
	BGPPath      *BGPPath
	FIBPath      *FIBPath
}

// Select returns negative if p < q, 0 if paths are equal, positive if p > q
func (p *Path) Select(q *Path) int8 {
	switch {
	case p == nil && q == nil:
		return 0
	case p == nil:
		return -1
	case q == nil:
		return 1
	default:
	}

	if p.Type > q.Type {
		return 1
	}

	if p.Type < q.Type {
		return -1
	}

	switch p.Type {
	case BGPPathType:
		return p.BGPPath.Select(q.BGPPath)
	case StaticPathType:
		return p.StaticPath.Select(q.StaticPath)
	case FIBPathType:
		return p.FIBPath.Select(q.FIBPath)
	}

	return 0
}

// ECMP checks if path p and q are equal enough to be considered for ECMP usage
func (p *Path) ECMP(q *Path) bool {
	switch p.Type {
	case BGPPathType:
		return p.BGPPath.ECMP(q.BGPPath)
	case StaticPathType:
		return p.StaticPath.ECMP(q.StaticPath)
	case FIBPathType:
		return p.FIBPath.ECMP(q.FIBPath)
	}

	panic("Unknown path type")
}

// ToProto converts path to proto path
func (p *Path) ToProto() *api.Path {
	a := &api.Path{
		StaticPath:  p.StaticPath.ToProto(),
		BgpPath:     p.BGPPath.ToProto(),
		TimeLearned: p.LTime,
	}

	switch p.Type {
	case StaticPathType:
		a.Type = api.Path_Static
	case BGPPathType:
		a.Type = api.Path_BGP
	}

	switch p.HiddenReason {
	case HiddenReasonNone:
		a.HiddenReason = api.Path_HiddenReasonNone
	case HiddenReasonNextHopUnreachable:
		a.HiddenReason = api.Path_HiddenReasonNextHopUnreachable
	case HiddenReasonFilteredByPolicy:
		a.HiddenReason = api.Path_HiddenReasonFilteredByPolicy
	case HiddenReasonASLoop:
		a.HiddenReason = api.Path_HiddenReasonASLoop
	case HiddenReasonOurOriginatorID:
		a.HiddenReason = api.Path_HiddenReasonOurOriginatorID
	case HiddenReasonClusterLoop:
		a.HiddenReason = api.Path_HiddenReasonClusterLoop
	case HiddenReasonOTCMismatch:
		a.HiddenReason = api.Path_HiddenReasonOTCMismatch
	}

	return a
}

// Compare checks if paths p and q are the same
func (p *Path) Compare(q *Path) bool {
	if p == nil || q == nil {
		return false
	}

	if p.Type != q.Type {
		return false
	}

	switch p.Type {
	case BGPPathType:
		return p.BGPPath.Compare(q.BGPPath)
	case StaticPathType:
		return p.StaticPath.Compare(q.StaticPath)
	}

	return false
}

// Equal checks if paths p and q are equal
func (p *Path) Equal(q *Path) bool {
	if p == nil || q == nil {
		return false
	}

	if p.Type != q.Type {
		return false
	}

	switch p.Type {
	case BGPPathType:
		return p.BGPPath.Equal(q.BGPPath)
	case StaticPathType:
		return p.StaticPath.Equal(q.StaticPath)
	}

	return p.Select(q) == 0
}

// PathsDiff gets the list of elements contained by a but not b
func PathsDiff(a, b []*Path) []*Path {
	ret := make([]*Path, 0)

	for _, pa := range a {
		if !pathsContains(pa, b) {
			ret = append(ret, pa)
		}
	}

	return ret
}

func pathsContains(needle *Path, haystack []*Path) bool {
	for _, p := range haystack {
		if p == needle {
			return true
		}
	}

	return false
}

// Print all known information about a route in logfile friendly format
func (p *Path) String() string {
	switch p.Type {
	case StaticPathType:
		return "not implemented yet"
	case BGPPathType:
		return p.BGPPath.String()
	case FIBPathType:
		return p.FIBPath.String()
	default:
		return fmt.Sprintf("Unknown path type. Probably not implemented yet (%d)", p.Type)
	}
}

// Print all known information about a route in human readable form
func (p *Path) Print() string {
	buf := &strings.Builder{}

	protocol := ""
	switch p.Type {
	case StaticPathType:
		protocol = "static"
	case BGPPathType:
		protocol = "BGP"
	case FIBPathType:
		protocol = "Netlink"
	}

	fmt.Fprintf(buf, "\tProtocol: %s\n", protocol)

	hr := p.HiddenReasonString()
	if hr != "" {
		fmt.Fprintf(buf, "\tHidden Reason: %s\n", hr)
	}

	if p.LTime != 0 {
		fmt.Fprintf(buf, "\tAge: %s\n", time.Since(time.Unix(int64(p.LTime), 0)).Truncate(time.Second).String())

	}

	switch p.Type {
	case StaticPathType:
		buf.WriteString("Not implemented yet")
	case BGPPathType:
		buf.WriteString(p.BGPPath.Print())
	case FIBPathType:
		buf.WriteString(p.FIBPath.Print())
	}

	return buf.String()
}

// Copy a route
func (p *Path) Copy() *Path {
	if p == nil {
		return nil
	}

	cp := *p
	cp.BGPPath = cp.BGPPath.Copy()
	cp.StaticPath = cp.StaticPath.Copy()

	return &cp
}

// NextHop returns the next hop IP Address
func (p *Path) NextHop() *bnet.IP {
	switch p.Type {
	case BGPPathType:
		return p.BGPPath.BGPPathA.NextHop
	case StaticPathType:
		return p.StaticPath.NextHop
	case FIBPathType:
		return p.FIBPath.NextHop
	}

	panic("Unknown path type")
}

// IsHidden returns if the path is hidden
func (p *Path) IsHidden() bool {
	return p.HiddenReason != HiddenReasonNone
}

// HiddenReasonString returns a human readable reason why this path is hidden (if any)
func (p *Path) HiddenReasonString() string {
	switch p.HiddenReason {
	case HiddenReasonNone:
		return ""
	case HiddenReasonNextHopUnreachable:
		return "Next-Hop unreachable"
	case HiddenReasonFilteredByPolicy:
		return "Filtered by Policy"
	case HiddenReasonASLoop:
		return "AS Path loop"
	case HiddenReasonOurOriginatorID:
		return "Found our Router ID as Originator ID"
	case HiddenReasonClusterLoop:
		return "Found our cluster ID in cluster list"
	case HiddenReasonOTCMismatch:
		return "OTC mismatch"
	default:
		return "unknown"
	}
}
