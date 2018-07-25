package netlink

import "github.com/vishvananda/netlink"

// RoutesDiff gets the list of elements contained by a but not b
func RoutesDiff(a, b []netlink.Route) []netlink.Route {
	ret := make([]netlink.Route, 0)

	for _, pa := range a {
		if !routesContains(pa, b) {
			ret = append(ret, pa)
		}
	}

	return ret
}

func routesContains(needle netlink.Route, haystack []netlink.Route) bool {
	for _, p := range haystack {
		if p.LinkIndex == needle.LinkIndex &&
			p.ILinkIndex == needle.ILinkIndex &&
			p.Scope == needle.Scope &&
			p.Dst == needle.Dst &&
			p.Protocol == needle.Protocol &&
			p.Priority == needle.Priority &&
			p.Table == needle.Table &&
			p.Type == needle.Type &&
			p.Tos == needle.Tos &&
			p.Flags == needle.Flags &&
			p.MPLSDst == needle.MPLSDst &&
			p.NewDst == needle.NewDst &&
			p.Encap == needle.Encap &&
			p.MTU == needle.MTU &&
			p.AdvMSS == needle.AdvMSS {

			return true
		}
	}

	return false
}
