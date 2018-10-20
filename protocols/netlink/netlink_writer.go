package protocolnetlink

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// NetlinkWriter is a locRIB subscriber which serializes routes from the locRIB to the Linux Kernel routing stack
type NetlinkWriter struct {
	options *config.Netlink
	filter  *filter.Filter

	// Routingtable for buffering, to ensure no double writes (a.k.a rtnetlink: file exists)
	mu sync.RWMutex
	pt map[bnet.Prefix][]*route.Path
}

// NewNetlinkWriter creates a new NetlinkWriter object and returns the pointer to it
func NewNetlinkWriter(options *config.Netlink) *NetlinkWriter {
	return &NetlinkWriter{
		options: options,
		filter:  options.ExportFilter,
		pt:      make(map[bnet.Prefix][]*route.Path),
	}
}

// UpdateNewClient Not supported for NetlinkWriter, since the writer is not observable
func (nw *NetlinkWriter) UpdateNewClient(routingtable.RouteTableClient) error {
	return fmt.Errorf("Not supported")
}

// Register Not supported for NetlinkWriter, since the writer is not observable
func (nw *NetlinkWriter) Register(routingtable.RouteTableClient) {
	log.Error("Not supported")
}

// RegisterWithOptions Not supported, since the writer is not observable
func (nw *NetlinkWriter) RegisterWithOptions(routingtable.RouteTableClient, routingtable.ClientOptions) {
	log.Error("Not supported")
}

// Unregister is not supported, since the writer is not observable
func (nw *NetlinkWriter) Unregister(routingtable.RouteTableClient) {
	log.Error("Not supported")
}

// RouteCount returns the number of stored routes
func (nw *NetlinkWriter) RouteCount() int64 {
	nw.mu.RLock()
	defer nw.mu.RUnlock()
	return int64(len(nw.pt))
}

// AddPath adds a path to the Kernel using netlink. This function is triggered by the loc_rib, cause we are subscribed as client in the loc_rib
func (nw *NetlinkWriter) AddPath(pfx bnet.Prefix, path *route.Path) error {
	// check if for this prefix already a route is existing
	existingPaths, ok := nw.pt[pfx]

	// if no route exists, add that route
	if existingPaths == nil || !ok {
		paths := make([]*route.Path, 1)
		paths = append(paths, path)
		nw.pt[pfx] = paths

		// add the route to kernel
		return nw.addKernel(pfx, path)
	}

	// if the new path is already in, don't do anything
	for _, ePath := range existingPaths {
		if ePath.Equal(path) {
			return nil
		}
	}

	existingPaths = append(existingPaths, path)
	nw.pt[pfx] = existingPaths

	// now add to netlink
	return nw.addKernel(pfx, path)
}

// RemovePath removes a path from the Kernel using netlink This function is triggered by the loc_rib, cause we are subscribed as client in the loc_rib
func (nw *NetlinkWriter) RemovePath(pfx bnet.Prefix, path *route.Path) bool {
	// check if for this prefix already a route is existing
	existingPaths, ok := nw.pt[pfx]

	// if no route exists, nothing to do
	if existingPaths == nil || !ok {
		return true
	}

	// if the new path is already in: remove
	removeIdx := 0
	remove := false
	for idx, ePath := range existingPaths {
		if ePath.Equal(path) {
			removeIdx = idx

			remove = true
			err := nw.removeKernel(pfx, path)
			if err != nil {
				log.WithError(err).Errorf("Error while removing path %s for prefix %s", path.String(), pfx.String())
				remove = false
			}

			break
		}
	}

	if remove {
		existingPaths = append(existingPaths[:removeIdx], existingPaths[removeIdx+1:]...)
		nw.pt[pfx] = existingPaths
	}

	return true
}

// Add pfx/path to kernel
func (nw *NetlinkWriter) addKernel(pfx bnet.Prefix, path *route.Path) error {
	route, err := nw.createRoute(pfx, path)
	if err != nil {
		log.Errorf("Error while creating route: %v", err)
		return fmt.Errorf("Error while creating route: %v", err)
	}

	log.WithFields(log.Fields{
		"Prefix": pfx.String(),
		"Table":  route.Table,
	}).Debug("AddPath to netlink")

	err = netlink.RouteAdd(route)
	if err != nil {
		log.Errorf("Error while adding route: %v", err)
		return fmt.Errorf("Error while adding route: %v", err)
	}

	return nil
}

// remove pfx/path from kernel
func (nw *NetlinkWriter) removeKernel(pfx bnet.Prefix, path *route.Path) error {
	log.WithFields(log.Fields{
		"Prefix": pfx.String(),
	}).Debug("Remove from netlink")

	route, err := nw.createRoute(pfx, path)
	if err != nil {
		return fmt.Errorf("Error while creating route: %v", err)
	}

	err = netlink.RouteDel(route)
	if err != nil {
		return fmt.Errorf("Error while removing route: %v", err)
	}

	return nil
}

// create a route from a prefix and a path
func (nw *NetlinkWriter) createRoute(pfx bnet.Prefix, path *route.Path) (*netlink.Route, error) {
	if path.Type != route.NetlinkPathType {
	}

	switch path.Type {
	case route.NetlinkPathType:
		return nw.createRouteFromNetlink(pfx, path)

	case route.BGPPathType:
		return nw.createRouteFromBGPPath(pfx, path)

	default:
		return nil, fmt.Errorf("PathType %d is not supported for adding to netlink", path.Type)
	}
}

func (nw *NetlinkWriter) createRouteFromNetlink(pfx bnet.Prefix, path *route.Path) (*netlink.Route, error) {
	nlPath := path.NetlinkPath

	log.WithFields(log.Fields{
		"Dst":      nlPath.Dst,
		"Src":      nlPath.Src,
		"NextHop":  nlPath.NextHop,
		"Priority": nlPath.Priority,
		"Protocol": nlPath.Protocol,
		"Type":     nlPath.Type,
		"Table":    nw.options.RoutingTable,
	}).Debug("created route")

	return &netlink.Route{
		Dst:      nlPath.Dst.GetIPNet(),
		Src:      nlPath.Src.Bytes(),
		Gw:       nlPath.NextHop.Bytes(),
		Priority: nlPath.Priority,
		Type:     nlPath.Type,
		Table:    nw.options.RoutingTable, // config dependent
		Protocol: route.ProtoBio,          // fix
	}, nil
}

func (nw *NetlinkWriter) createRouteFromBGPPath(pfx bnet.Prefix, path *route.Path) (*netlink.Route, error) {
	bgpPath := path.BGPPath

	log.WithFields(log.Fields{
		"Dst":           pfx,
		"NextHop":       bgpPath.NextHop,
		"Protocol":      "BGP",
		"BGPIdentifier": bgpPath.BGPIdentifier,
		"Table":         nw.options.RoutingTable,
	}).Debug("created route")

	return &netlink.Route{
		Dst:      pfx.GetIPNet(),
		Gw:       bgpPath.NextHop.Bytes(),
		Table:    nw.options.RoutingTable, // config dependent
		Protocol: route.ProtoBio,          // fix
	}, nil

}
