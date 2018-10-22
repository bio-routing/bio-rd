package protocolnetlink

import (
	"fmt"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// Constants for IP family
const (
	IPFamily4 = 4 // IPv4
	IPFamily6 = 6 // IPv6
)

// NetlinkReader read routes from the Linux Kernel and propagates it to the locRIB
type NetlinkReader struct {
	options *config.Netlink
	routingtable.ClientManager
	filter *filter.Filter

	mu     sync.RWMutex
	routes []netlink.Route
}

// NewNetlinkReader creates a new reader object and returns the pointer to it
func NewNetlinkReader(options *config.Netlink) *NetlinkReader {
	nr := &NetlinkReader{
		options: options,
		filter:  options.ImportFilter,
	}

	nr.ClientManager = routingtable.NewClientManager(nr)

	return nr
}

// Read reads routes from the kernel
func (nr *NetlinkReader) Read() {
	log.WithField("rt_table", nr.options.RoutingTable).Info("Started netlink server")

	// Start fetching the kernel routes after the hold time
	time.Sleep(nr.options.HoldTime)

	for {
		// Family doesn't matter. I only filter by the rt_table here
		routes, err := netlink.RouteListFiltered(IPFamily4, &netlink.Route{Table: nr.options.RoutingTable}, netlink.RT_FILTER_TABLE)
		if err != nil {
			log.WithError(err).Panic("Failed to read routes from kernel")
		}

		nr.propagateChanges(routes)

		nr.mu.Lock()
		nr.routes = routes

		log.Debugf("NetlinkRouteDiff: %d", len(route.NetlinkRouteDiff(nr.routes, routes)))
		nr.mu.Unlock()

		time.Sleep(nr.options.UpdateInterval)
	}
}

// create a path from a route
func createPathFromRoute(r *netlink.Route) (*route.Path, error) {
	nlPath, err := route.NewNlPathFromRoute(r, true)

	if err != nil {
		return nil, fmt.Errorf("Error while creating path object from route object: %v", err)
	}

	return &route.Path{
		Type:        route.NetlinkPathType,
		NetlinkPath: nlPath,
	}, nil
}

// propagate changes to all subscribed clients
func (nr *NetlinkReader) propagateChanges(routes []netlink.Route) {
	nr.removePathsFromClients(routes)
	nr.addPathsToClients(routes)
}

// Add given paths to clients
func (nr *NetlinkReader) addPathsToClients(routes []netlink.Route) {
	// only advertise changed routes
	nr.mu.RLock()
	advertise := route.NetlinkRouteDiff(routes, nr.routes)
	nr.mu.RUnlock()

	for _, r := range advertise {
		// Is it a BIO-Written route? if so, skip it, dont advertise it
		if r.Protocol == route.ProtoBio {
			log.WithFields(routeLogFields(r)).Debug("Skipping bio route")
			continue
		}

		// create pfx and path from route
		pfx := bnet.NewPfxFromIPNet(r.Dst)
		path, err := createPathFromRoute(&r)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"prefix": pfx.String(),
				"path":   path.String(),
			}).Error("Unable to create path")
			continue
		}

		// Apply filter (if existing)
		if nr.filter != nil {
			var reject bool
			// TODO: Implement filter that cann handle netlinkRoute objects
			path, reject = nr.filter.ProcessTerms(pfx, path)
			if reject {
				log.WithError(err).WithFields(log.Fields{
					"prefix": pfx.String(),
					"path":   path.String(),
				}).Debug("Skipping route due to filter")

				continue
			}
		}

		for _, client := range nr.ClientManager.Clients() {
			log.WithFields(log.Fields{
				"pfx":  pfx,
				"path": path,
			}).Debug("NetlinkReader - client.AddPath")
			client.AddPath(pfx, path)
		}
	}
}

// Remove given paths from clients
func (nr *NetlinkReader) removePathsFromClients(routes []netlink.Route) {
	nr.mu.RLock()

	// get the number of routes
	routeLength := len(nr.routes)

	// If there where no routes yet, just skip this funktion. There's nothing to delete
	if routeLength == 0 {
		nr.mu.RUnlock()
		return
	}

	// only withdraw changed routes
	withdraw := route.NetlinkRouteDiff(nr.routes, routes)
	nr.mu.RUnlock()

	for _, r := range withdraw {
		// Is it a BIO-Written route? if so, skip it, dont advertise it
		if r.Protocol == route.ProtoBio {
			continue
		}

		// create pfx and path from route
		pfx := bnet.NewPfxFromIPNet(r.Dst)
		path, err := createPathFromRoute(&r)
		if err != nil {
			log.WithError(err).Error("Unable to create path")
			continue
		}

		// Apply filter (if existing)
		if nr.filter != nil {
			var reject bool
			// TODO: Implement filter that cann handle netlinkRoute objects
			path, reject = nr.filter.ProcessTerms(pfx, path)
			if reject {
				continue
			}
		}

		for _, client := range nr.ClientManager.Clients() {
			log.WithFields(log.Fields{
				"pfx":  pfx,
				"path": path,
			}).Debug("NetlinkReader - client.RemovePath")
			client.RemovePath(pfx, path)
		}
	}
}

func routeLogFields(route netlink.Route) log.Fields {
	return log.Fields{
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
	}
}

// AddPath is Not supported
func (nr *NetlinkReader) AddPath(bnet.Prefix, *route.Path) error {
	return fmt.Errorf("Not supported")
}

// RemovePath is Not supported
func (nr *NetlinkReader) RemovePath(bnet.Prefix, *route.Path) bool {
	return false
}

// UpdateNewClient is currently not supported
func (nr *NetlinkReader) UpdateNewClient(routingtable.RouteTableClient) error {
	return fmt.Errorf("Not supported")
}

// Register is currently not supported
func (nr *NetlinkReader) Register(routingtable.RouteTableClient) {
}

// RegisterWithOptions is Not supported
func (nr *NetlinkReader) RegisterWithOptions(routingtable.RouteTableClient, routingtable.ClientOptions) {
}

// Unregister is Not supported
func (nr *NetlinkReader) Unregister(routingtable.RouteTableClient) {
}

// RouteCount retuns the number of routes stored in the internal routing table
func (nr *NetlinkReader) RouteCount() int64 {
	nr.mu.RLock()
	defer nr.mu.RUnlock()

	return int64(len(nr.routes))
}
