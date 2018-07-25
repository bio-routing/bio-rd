package netlink

import (
	"fmt"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	rt "github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

	bnet "github.com/bio-routing/bio-rd/net"
)

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

type NetlinkServer struct {
	mu      sync.RWMutex
	rib     *locRIB.LocRIB
	options *config.Netlink
	routes  []netlink.Route
	routingtable.ClientManager
}

func NewNetlinkServer(options *config.Netlink, rib *locRIB.LocRIB) *NetlinkServer {
	n := &NetlinkServer{
		options: options,
		rib:     rib,
	}

	n.ClientManager = routingtable.NewClientManager(n)

	return n
}

func (n *NetlinkServer) Start() error {
	log.WithField("rt_table", n.options.RoutingTable).Info("Started netlink server")

	n.rib.ClientManager.Register(n)
	go n.readKernelRoutes()

	return nil
}

// Register a new client
func (n *NetlinkServer) Register(client rt.RouteTableClient) {
	n.ClientManager.Register(client)
}

// Register a new client with options
func (n *NetlinkServer) RegisterWithOptions(client rt.RouteTableClient, options rt.ClientOptions) {
	n.ClientManager.RegisterWithOptions(client, options)
}

// Unregister a client
func (n *NetlinkServer) Unregister(client rt.RouteTableClient) {
	n.ClientManager.Unregister(client)
}

// UpdateNewClient sends current state to a new client
func (n *NetlinkServer) UpdateNewClient(client routingtable.RouteTableClient) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	n.propagateChanges(make([]netlink.Route, 0), n.routes)

	return nil
}

// Get the route count
func (n *NetlinkServer) RouteCount() int64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return int64(len(n.routes))
}

// Add a path to the Kernel using netlink
// This function is triggered by the loc_rib, cause we are subscribed as
// client in the loc_rib
func (n *NetlinkServer) AddPath(pfx net.Prefix, path *route.Path) error {
	log.WithFields(log.Fields{
		"Prefix": pfx.String(),
	}).Debug("AddPath to netlink")

	err := netlink.RouteAdd(n.createRoute(pfx, path))
	if err != nil {
		log.WithError(err).Error("Error while adding route")
		return fmt.Errorf("Error while adding route: %v", err)
	}

	return nil
}

// Remove a path from the Kernel using netlink
// This function is triggered by the loc_rib, cause we are subscribed as
// client in the loc_rib
func (n *NetlinkServer) RemovePath(pfx net.Prefix, path *route.Path) bool {
	log.WithFields(log.Fields{
		"Prefix": pfx.String(),
	}).Debug("Remove from netlink")

	err := netlink.RouteDel(n.createRoute(pfx, path))
	if err != nil {
		log.WithError(err).Error("Error while removing route")
		return false
	}

	return true
}

// Read routes from kernel
func (n *NetlinkServer) readKernelRoutes() {
	// Start fetching the kernel routes after the hold time
	time.Sleep(n.options.HoldTime)

	for {
		// Family doesn't matter. I only filter by the rt_table here
		routes, err := netlink.RouteListFiltered(4, &netlink.Route{Table: n.options.RoutingTable}, netlink.RT_FILTER_TABLE)
		if err != nil {
			log.WithError(err).Error("Failed to read routes from kernel")
			// TODO: what should happen in this case? wait a few seconds and continue? break?
			break
		}

		n.mu.Lock()
		n.propagateChanges(n.routes, routes)
		n.routes = routes
		n.mu.Unlock()

		n.mu.RLock()
		for idx, route := range n.routes {
			log.WithFields(log.Fields{
				"LinkIndex":  route.LinkIndex,
				"ILinkIndex": route.ILinkIndex,
				"Scope":      route.Scope,
				"Dst":        route.Dst,
				"Src":        route.Src,
				"Gw":         route.Gw,
				"MultiPath":  route.MultiPath,
				"Protocol":   route.Protocol,
				"Priority":   route.Priority,
				"Table":      route.Table,
				"Type":       route.Type,
				"Tos":        route.Tos,
				"Flags":      route.Flags,
				"MPLSDst":    route.MPLSDst,
				"NewDst":     route.NewDst,
				"Encap":      route.Encap,
				"MTU":        route.MTU,
				"AdvMSS":     route.AdvMSS,
			}).Debugf("Route [%d] read", idx)
		}
		log.WithField("RouteCount", n.RouteCount()).Debugf("Reading routes finised")
		n.mu.RUnlock()

		time.Sleep(n.options.UpdateInterval)
	}
}

// create a route from a prefix and a path
func (n *NetlinkServer) createRoute(pfx net.Prefix, path *route.Path) *netlink.Route {
	var nextHop bnet.IP

	switch path.Type {
	case route.BGPPathType:
		nextHop = path.BGPPath.NextHop
	case route.StaticPathType:
		nextHop = path.StaticPath.NextHop
	}

	log.WithFields(log.Fields{
		"Dst":   pfx.GetIPNet(),
		"Gw":    nextHop.Bytes(),
		"Table": n.options.RoutingTable,
	}).Debug("created route")

	return &netlink.Route{
		Dst:   pfx.GetIPNet(),
		Gw:    nextHop.Bytes(),
		Table: n.options.RoutingTable,
	}
}

// create a path from a route
func createPathFromRoute(r *netlink.Route) *route.Path {
	nextHop, _ := net.IPFromBytes(r.Dst.IP)
	return &route.Path{
		Type:       route.StaticPathType,
		StaticPath: &route.StaticPath{NextHop: nextHop},
	}
}

// propagate changes to all subscribed clients
func (n *NetlinkServer) propagateChanges(oldRoutes []netlink.Route, newRoutes []netlink.Route) {
	n.removePathsFromClients(oldRoutes, newRoutes)
	n.addPathsToClients(oldRoutes, newRoutes)
}

func (n *NetlinkServer) addPathsToClients(oldRoutes []netlink.Route, newRoutes []netlink.Route) {
	for _, client := range n.ClientManager.Clients() {
		advertise := RoutesDiff(newRoutes, oldRoutes)

		for _, route := range advertise {
			pfx := net.NewPfxFromIPNet(route.Dst)
			path := createPathFromRoute(&route)
			client.AddPath(pfx, path)
		}
	}
}

func (n *NetlinkServer) removePathsFromClients(oldRoutes []netlink.Route, newRoutes []netlink.Route) {
	for _, client := range n.ClientManager.Clients() {
		withdraw := RoutesDiff(oldRoutes, newRoutes)

		for _, route := range withdraw {
			pfx := net.NewPfxFromIPNet(route.Dst)
			path := createPathFromRoute(&route)
			client.RemovePath(pfx, path)
		}
	}
}
