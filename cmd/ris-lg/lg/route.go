package lg

import (
	"github.com/bio-routing/bio-rd/protocols/bgp/types"

	bnet "github.com/bio-routing/bio-rd/net"
	routeAPI "github.com/bio-routing/bio-rd/route/api"
)

type Route struct {
	Prefix string
	Paths  []Path
}

type Path struct {
	Source    string
	Nexthop   string
	LocalPref uint32
	ASPath    string
	MED       uint32
}

func convertRoutes(routes []*routeAPI.Route) []Route {
	ret := make([]Route, 0, len(routes))
	for _, r := range routes {
		ret = append(ret, convertRoute(r))
	}

	return ret
}

func convertRoute(r *routeAPI.Route) Route {
	ret := Route{
		Prefix: bnet.NewPrefixFromProtoPrefix(r.Pfx).String(),
		Paths:  make([]Path, 0, len(r.Paths)),
	}

	for _, p := range r.Paths {
		if p.BgpPath == nil {
			continue
		}

		q := Path{
			Source:    bnet.IPFromProtoIP(p.BgpPath.Source).String(),
			Nexthop:   bnet.IPFromProtoIP(p.BgpPath.NextHop).String(),
			LocalPref: p.BgpPath.LocalPref,
			ASPath:    types.ASPathFromProtoASPath(p.BgpPath.AsPath).String(),
			MED:       p.BgpPath.Med,
		}

		ret.Paths = append(ret.Paths, q)

	}

	return ret
}
