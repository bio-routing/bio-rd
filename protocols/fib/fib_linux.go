package fib

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/vishvananda/netlink"
)

func (f *FIB) loadFIB() {
	f.osAdapter = newOSFIBLinux(f)
}

type osFibAdapterLinux struct {
	fib *FIB
}

func newOSFIBLinux(f *FIB) *osFibAdapterLinux {
	fib := &osFibAdapterLinux{
		fib: f,
	}

	return fib
}

func (fib *osFibAdapterLinux) addPath(path route.FIBPath) error {
	return fmt.Errorf("Not implemented")
}

func (fib *osFibAdapterLinux) removePath(path route.FIBPath) error {
	return fmt.Errorf("Not implemented")
}

func (fib *osFibAdapterLinux) pathCount() int64 {
	return 0
}

func (fib *osFibAdapterLinux) start() error {
	return fmt.Errorf("Not implemented")
}

// NetlinkRouteDiff gets the list of elements contained by a but not b
func NetlinkRouteDiff(a, b []netlink.Route) []netlink.Route {
	ret := make([]netlink.Route, 0)

	for _, pa := range a {
		if !netlinkRoutesContains(pa, b) {
			ret = append(ret, pa)
		}
	}

	return ret
}

func netlinkRoutesContains(needle netlink.Route, haystack []netlink.Route) bool {
	for i := range haystack {
		if netlinkRouteEquals(&needle, &haystack[i]) {
			return true
		}
	}

	return false
}

func netlinkRouteEquals(a, b *netlink.Route) bool {
	aMaskSize, aMaskBits := a.Dst.Mask.Size()
	bMaskSize, bMaskBits := b.Dst.Mask.Size()

	return a.LinkIndex == b.LinkIndex &&
		a.ILinkIndex == b.ILinkIndex &&
		a.Scope == b.Scope &&

		a.Dst.IP.Equal(b.Dst.IP) &&
		aMaskSize == bMaskSize &&
		aMaskBits == bMaskBits &&

		a.Src.Equal(b.Src) &&
		a.Gw.Equal(b.Gw) &&

		a.Protocol == b.Protocol &&
		a.Priority == b.Priority &&
		a.Table == b.Table &&
		a.Type == b.Type &&
		a.Tos == b.Tos &&
		a.Flags == b.Flags &&
		a.MTU == b.MTU &&
		a.AdvMSS == b.AdvMSS
}

// NewPathsFromNlRoute creates a new route.FIBPath object from a netlink.Route object
func NewPathsFromNlRoute(r netlink.Route, kernel bool) (bnet.Prefix, []*route.Path, error) {
	var src bnet.IP
	var dst bnet.Prefix

	if r.Src == nil && r.Dst == nil {
		return bnet.Prefix{}, nil, fmt.Errorf("Cannot create NlPath, since source and destination are both nil")
	}

	if r.Src == nil && r.Dst != nil {
		dst = bnet.NewPfxFromIPNet(r.Dst)
		if dst.Addr().IsIPv4() {
			src = bnet.IPv4(0)
		} else {
			src = bnet.IPv6(0, 0)
		}
	}

	if r.Src != nil && r.Dst == nil {
		src, _ = bnet.IPFromBytes(r.Src)
		if src.IsIPv4() {
			dst = bnet.NewPfx(bnet.IPv4(0), 0)
		} else {
			dst = bnet.NewPfx(bnet.IPv6(0, 0), 0)
		}
	}

	if r.Src != nil && r.Dst != nil {
		src, _ = bnet.IPFromBytes(r.Src)
		dst = bnet.NewPfxFromIPNet(r.Dst)
	}

	paths := make([]*route.Path, 0)

	if len(r.MultiPath) > 0 {
		for _, multiPath := range r.MultiPath {
			nextHop, _ := bnet.IPFromBytes(multiPath.Gw)
			paths = append(paths, &route.Path{
				Type: route.FIBPathType,
				FIBPath: &route.FIBPath{
					Src:      src,
					NextHop:  nextHop,
					Priority: r.Priority,
					Protocol: r.Protocol,
					Type:     r.Type,
					Table:    r.Table,
					Kernel:   kernel,
				},
			})
		}
	} else {
		nextHop, _ := bnet.IPFromBytes(r.Gw)
		paths = append(paths, &route.Path{
			Type: route.FIBPathType,
			FIBPath: &route.FIBPath{
				Src:      src,
				NextHop:  nextHop,
				Priority: r.Priority,
				Protocol: r.Protocol,
				Type:     r.Type,
				Table:    r.Table,
				Kernel:   kernel,
			},
		})
	}

	return dst, paths, nil
}
