package netlink

import (
	"fmt"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	rt "github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"

	bnet "github.com/bio-routing/bio-rd/net"
)

type NetlinkServer struct {
	mu      sync.RWMutex
	rib     *locRIB.LocRIB
	options *config.Netlink
	routes  []netlink.Route
}

func NewNetlinkServer(options *config.Netlink, rib *locRIB.LocRIB) *NetlinkServer {
	return &NetlinkServer{
		options: options,
		rib:     rib,
	}
}

func (n *NetlinkServer) Start() error {
	log.WithField("rt_table", n.options.RoutingTable).Info("Started netlink server")

	n.rib.ClientManager.Register(n)
	go n.readKernelRoutes()

	return nil
}

func (n *NetlinkServer) RouteCount() int64 {
	return int64(len(n.routes))
}
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

func (n *NetlinkServer) readKernelRoutes() {
	// Start fetching the kernel routes after the hold time
	time.Sleep(n.options.HoldTime)

	for {

		// Family doesn't matter. I only filter by the rt_table here
		routes, err := netlink.RouteListFiltered(4, &netlink.Route{Table: n.options.RoutingTable}, netlink.RT_FILTER_TABLE)
		if err != nil {
			log.WithError(err).Error("Failed to read routes from kernel")
			// TODO: what should happen in this case? continue? break? Since it's a daemon i would say wait...
		}
		n.mu.Lock()
		n.routes = routes
		n.mu.Unlock()

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

		time.Sleep(n.options.UpdateInterval)
	}
}

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

// Not supported for netlink
func (n *NetlinkServer) Register(rt.RouteTableClient) {
}

// Not supported for netlink
func (n *NetlinkServer) RegisterWithOptions(rt.RouteTableClient, rt.ClientOptions) {
}

// Not supported for netlink
func (n *NetlinkServer) Unregister(rt.RouteTableClient) {
}

// Not supported for netlink
func (n *NetlinkServer) UpdateNewClient(rt.RouteTableClient) error {
	return nil
}
