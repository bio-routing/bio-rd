package fib

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func (f *FIB) loadOSAdapter() {
	f.osAdapter = newOSFIBLinux(f)
}

type osFibAdapterLinux struct {
	fib *FIB
}

func newOSFIBLinux(f *FIB) *osFibAdapterLinux {
	linuxAdapter := &osFibAdapterLinux{
		fib: f,
	}

	return linuxAdapter
}

func (f *osFibAdapterLinux) addPath(pfx bnet.Prefix, paths []*route.FIBPath) error {
	route, err := f.createRoute(pfx, paths)
	if err != nil {
		return errors.Wrap(err, "Could not create route from prefix and path: %v")
	}

	log.WithFields(log.Fields{
		"Prefix": pfx.String(),
		"Route":  route,
	}).Debug("AddPath to netlink")

	err = netlink.RouteAdd(route)
	if err != nil && err.Error() != "file exists" {
		return errors.Wrap(err, "Error while adding route")
	}

	return nil
}

func (f *osFibAdapterLinux) removePath(pfx bnet.Prefix, path *route.FIBPath) error {
	nlRoute, err := f.createRoute(pfx, []*route.FIBPath{path})
	if err != nil {
		return errors.Wrap(err, "Could not create route from prefix and path: %v")
	}

	log.WithFields(log.Fields{
		"Prefix": pfx.String(),
	}).Debug("Remove from netlink")

	err = netlink.RouteDel(nlRoute)
	if err != nil {
		return errors.Wrap(err, "Error while removing route")
	}

	return nil
}

// create a route from a prefix and a path
func (f *osFibAdapterLinux) createRoute(pfx bnet.Prefix, paths []*route.FIBPath) (*netlink.Route, error) {
	route := &netlink.Route{
		Dst:      pfx.GetIPNet(),
		Table:    int(f.fib.vrf.ID()),
		Protocol: route.ProtoBio,
	}

	multiPath := make([]*netlink.NexthopInfo, 0)

	for _, path := range paths {
		nextHop := &netlink.NexthopInfo{
			Gw: path.NextHop.Bytes(),
		}
		multiPath = append(multiPath, nextHop)
	}

	if len(multiPath) == 1 {
		route.Gw = multiPath[0].Gw
	} else if len(multiPath) > 1 {
		route.MultiPath = multiPath
	} else {
		return nil, fmt.Errorf("No destination address specified. At least one NextHop has to be specified in path")
	}

	return route, nil
}
