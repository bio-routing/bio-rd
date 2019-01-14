package fib

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/vishvananda/netlink"
)

const (
	ProtoUnspec   = 0  // unspec (from /etc/iproute2/rt_protos)
	ProtoRedirect = 1  // redirect (from /etc/iproute2/rt_protos)
	ProtoKernel   = 2  // kernel (from /etc/iproute2/rt_protos)
	ProtoBoot     = 3  // boot (from /etc/iproute2/rt_protos)
	ProtoStatic   = 4  // static (from /etc/iproute2/rt_protos)
	ProtoZebra    = 11 // zebra (from /etc/iproute2/rt_protos)
	ProtoBird     = 12 // bird (from /etc/iproute2/rt_protos)
	ProtoDHCP     = 16 // dhcp (from /etc/iproute2/rt_protos)
	ProtoBio      = 45 // bio
)

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

// NewNlPathFromRoute creates a new route.FIBPath object from a netlink.Route object
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
