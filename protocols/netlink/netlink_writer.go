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
	mu        sync.RWMutex
	pathTable map[bnet.Prefix][]*route.Path
}

// NewNetlinkWriter creates a new NetlinkWriter object and returns the pointer to it
func NewNetlinkWriter(options *config.Netlink) *NetlinkWriter {
	return &NetlinkWriter{
		options:   options,
		filter:    options.ExportFilter,
		pathTable: make(map[bnet.Prefix][]*route.Path),
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
	return int64(len(nw.pathTable))
}

// AddPath adds a path to the Kernel using netlink. This function is triggered by the loc_rib, cause we are subscribed as client in the loc_rib
func (nw *NetlinkWriter) AddPath(pfx bnet.Prefix, path *route.Path) error {
	// check if for this prefix already a route is existing
	existingPaths, ok := nw.pathTable[pfx]

	// if no route exists, add that route
	if existingPaths == nil || !ok {
		nw.pathTable[pfx] = []*route.Path{path}
		return nw.addKernel(pfx)
	}

	// if the new path is already in, don't do anything
	for _, ePath := range existingPaths {
		if ePath.Equal(path) {
			return nil
		}
	}

	// if newly added path is a ecmp path to the existing paths, add it
	if path.ECMP(existingPaths[0]) {
		nw.removeKernel(pfx, existingPaths)
		existingPaths = append(existingPaths, path)
		nw.pathTable[pfx] = existingPaths

		return nw.addKernel(pfx)
	}

	// if newly added path is no ecmp path to the existing ones, remove all old and only add the new
	nw.removeKernel(pfx, existingPaths)
	nw.pathTable[pfx] = []*route.Path{path}
	return nw.addKernel(pfx)

}

// RemovePath removes a path from the Kernel using netlink This function is triggered by the loc_rib, cause we are subscribed as client in the loc_rib
func (nw *NetlinkWriter) RemovePath(pfx bnet.Prefix, path *route.Path) bool {
	// check if for this prefix already a route is existing
	existingPaths, ok := nw.pathTable[pfx]

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
			err := nw.removeKernel(pfx, []*route.Path{path})
			if err != nil {
				log.WithError(err).Errorf("Error while removing path %s for prefix %s", path.String(), pfx.String())
				remove = false
			}

			break
		}
	}

	if remove {
		existingPaths = append(existingPaths[:removeIdx], existingPaths[removeIdx+1:]...)
		nw.pathTable[pfx] = existingPaths
	}

	return true
}

// Add pfx/path to kernel
func (nw *NetlinkWriter) addKernel(pfx bnet.Prefix) error {
	route := nw.createRoute(pfx, nw.pathTable[pfx])

	log.WithFields(log.Fields{
		"Prefix": pfx.String(),
		"Table":  route.Table,
	}).Debug("AddPath to netlink")

	err := netlink.RouteAdd(route)
	if err != nil {
		log.Errorf("Error while adding route: %v", err)
		return fmt.Errorf("Error while adding route: %v", err)
	}

	return nil
}

// remove pfx/path from kernel
func (nw *NetlinkWriter) removeKernel(pfx bnet.Prefix, paths []*route.Path) error {
	route := nw.createRoute(pfx, nw.pathTable[pfx])

	log.WithFields(log.Fields{
		"Prefix": pfx.String(),
	}).Debug("Remove from netlink")

	err := netlink.RouteDel(route)
	if err != nil {
		return fmt.Errorf("Error while removing route: %v", err)
	}

	return nil
}

// create a route from a prefix and a path
func (nw *NetlinkWriter) createRoute(pfx bnet.Prefix, paths []*route.Path) *netlink.Route {
	route := &netlink.Route{
		Dst:      pfx.GetIPNet(),
		Table:    int(nw.options.RoutingTable), // config dependent
		Protocol: route.ProtoBio,
	}

	multiPath := make([]*netlink.NexthopInfo, 0)

	for _, path := range paths {
		nextHop := &netlink.NexthopInfo{
			Gw: path.NextHop().Bytes(),
		}
		multiPath = append(multiPath, nextHop)
	}

	route.MultiPath = multiPath

	return route
}
