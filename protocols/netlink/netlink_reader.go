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
	IPFamily4 int = 4 // IPv4
	IPFamily6 int = 6 // IPv6
)

// NetlinkReader reads routes from the Linux Kernel and propagates them to the locRIB
type NetlinkReader struct {
	clientManager *routingtable.ClientManager
	options       *config.Netlink
	filter        *filter.Filter

	mu     sync.RWMutex
	routes []netlink.Route
}

// NewNetlinkReader creates a new reader object and returns the pointer to it
func NewNetlinkReader(options *config.Netlink) *NetlinkReader {
	nr := &NetlinkReader{
		options: options,
		filter:  options.ImportFilter,
	}

	nr.clientManager = routingtable.NewClientManager(nr)

	return nr
}

// Read reads routes from the kernel
func (nr *NetlinkReader) Read() {
	// Start fetching the kernel routes after the hold time
	time.Sleep(nr.options.HoldTime)

	for {
		// Family doesn't matter. I only filter by the rt_table here
		routes, err := netlink.RouteListFiltered(IPFamily4, &netlink.Route{Table: int(nr.options.RoutingTable)}, netlink.RT_FILTER_TABLE)
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

// propagate changes to all subscribed clients
func (nr *NetlinkReader) propagateChanges(routes []netlink.Route) {
	nr.removePathsFromClients(routes)
	nr.addPathsToClients(routes)
}

// Add given paths to clients
func (nr *NetlinkReader) addPathsToClients(routes []netlink.Route) {
	// If there were no routes yet, just skip this function. There's nothing to add
	if len(routes) == 0 {
		return
	}

	// only advertise changed routes
	nr.mu.RLock()
	advertise := route.NetlinkRouteDiff(routes, nr.routes)
	nr.mu.RUnlock()

	for _, client := range nr.clientManager.Clients() {
		for _, r := range advertise {
			if isBioRoute(r) {
				log.WithFields(routeLogFields(r)).Debug("Skipping bio route")
				continue
			}

			// create pfx and path from route
			pfx, paths, err := route.NewPathsFromNlRoute(r, true)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"prefix": pfx.String(),
				}).Error("Unable to create path")
				continue
			}

			for _, path := range paths {
				var p *route.Path
				if nr.filter != nil {
					var reject bool
					p, reject := nr.filter.ProcessTerms(pfx, path)
					if reject {
						log.WithError(err).WithFields(log.Fields{
							"prefix": pfx.String(),
							"path":   p.String(),
						}).Debug("Skipping route due to filter")
						continue
					}
				}

				log.WithFields(log.Fields{
					"pfx":  pfx,
					"path": p,
				}).Debug("NetlinkReader - client.AddPath")
				client.AddPath(pfx, p)
			}
		}
	}
}

// Remove given paths from clients
func (nr *NetlinkReader) removePathsFromClients(routes []netlink.Route) {
	nr.mu.RLock()

	// If there were no routes yet, just skip this function. There's nothing to delete
	if len(nr.routes) == 0 {
		nr.mu.RUnlock()
		return
	}

	// only withdraw changed routes
	withdraw := route.NetlinkRouteDiff(nr.routes, routes)
	nr.mu.RUnlock()

	for _, client := range nr.clientManager.Clients() {
		for _, r := range withdraw {
			if isBioRoute(r) {
				log.WithFields(routeLogFields(r)).Debug("Skipping bio route")
				continue
			}

			// create pfx and path from route
			pfx, paths, err := route.NewPathsFromNlRoute(r, true)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"prefix": pfx.String(),
				}).Error("Unable to create path")
				continue
			}

			for _, path := range paths {
				var p *route.Path
				if nr.filter != nil {
					var reject bool
					p, reject = nr.filter.ProcessTerms(pfx, path)
					if reject {
						log.WithError(err).WithFields(log.Fields{
							"prefix": pfx.String(),
							"path":   p.String(),
						}).Debug("Skipping route due to filter")
						continue
					}
				}

				log.WithFields(log.Fields{
					"pfx":  pfx,
					"path": p,
				}).Debug("NetlinkReader - client.RemovePath")
				client.RemovePath(pfx, p)
			}
		}
	}
}

// Is route a BIO-Written route?
func isBioRoute(r netlink.Route) bool {
	return uint32(r.Protocol) == route.ProtoBio
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

// Unregister is Not supported
func (nr *NetlinkReader) Unregister(routingtable.RouteTableClient) {
}

// RouteCount retuns the number of routes stored in the internal routing table
func (nr *NetlinkReader) RouteCount() int64 {
	nr.mu.RLock()
	defer nr.mu.RUnlock()

	return int64(len(nr.routes))
}
