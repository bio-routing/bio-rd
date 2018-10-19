package proto_netlink

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

type NetlinkReader struct {
	options *config.Netlink
	routingtable.ClientManager
	filter *filter.Filter

	mu     sync.RWMutex
	routes []netlink.Route
}

func NewNetlinkReader(options *config.Netlink) *NetlinkReader {
	nr := &NetlinkReader{
		options: options,
		filter:  options.ImportFilter,
	}

	nr.ClientManager = routingtable.NewClientManager(nr)

	return nr
}

// Read routes from kernel
func (nr *NetlinkReader) Read() {
	log.WithField("rt_table", nr.options.RoutingTable).Info("Started netlink server")

	// Start fetching the kernel routes after the hold time
	time.Sleep(nr.options.HoldTime)

	for {
		// Family doesn't matter. I only filter by the rt_table here
		routes, err := netlink.RouteListFiltered(4, &netlink.Route{Table: nr.options.RoutingTable}, netlink.RT_FILTER_TABLE)
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
		return nil, fmt.Errorf("Error while creating path object from route object", err)
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
	for _, client := range nr.ClientManager.Clients() {
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
				log.WithError(err).Error("Unable to create path")
				continue
			}

			// Apply filter (if existing)
			if nr.filter != nil {
				var reject bool
				// TODO: Implement filter that cann handle netlinkRoute objects
				path, reject = nr.filter.ProcessTerms(pfx, path)
				if reject {
					log.Debug("Skipping route due to filter")
					continue
				}
			}

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
	for _, client := range nr.ClientManager.Clients() {
		// If there where no routes yet, just skip this funktion. There's nothing to delete
		nr.mu.RLock()
		if len(nr.routes) == 0 {
			nr.mu.RUnlock()
			break
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

// Not supported
func (nr *NetlinkReader) AddPath(bnet.Prefix, *route.Path) error {
	return fmt.Errorf("Not supported")
}

// Not supported
func (nr *NetlinkReader) RemovePath(bnet.Prefix, *route.Path) bool {
	return false
}

// Not supported
func (nr *NetlinkReader) UpdateNewClient(routingtable.RouteTableClient) error {
	return fmt.Errorf("Not supported")
}

func (nr *NetlinkReader) Register(routingtable.RouteTableClient) {
}

func (nr *NetlinkReader) RegisterWithOptions(routingtable.RouteTableClient, routingtable.ClientOptions) {
}

func (nr *NetlinkReader) Unregister(routingtable.RouteTableClient) {
}

func (nr *NetlinkReader) RouteCount() int64 {
	nr.mu.RLock()
	defer nr.mu.RUnlock()

	return int64(len(nr.routes))
}
