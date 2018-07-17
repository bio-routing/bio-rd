package server

import (
	"fmt"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	rt "github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type netlinkServer struct {
	table int
	rib   *locRIB.LocRIB
}

func NewNetlinkServer(rib *locRIB.LocRIB) *netlinkServer {
	return &netlinkServer{
		table: 0,
		rib:   rib,
	}
}

func (n *netlinkServer) Start(c *config.Global) error {
	if err := c.SetDefaultGlobalConfigValues(); err != nil {
		return fmt.Errorf("Failed to load defaults: %v", err)
	}

	n.table = c.RoutingTable
	log.Infof("Kernel routing table: %d\n", n.table)

	n.rib.ClientManager.Register(n)

	return nil
}

func (n *netlinkServer) AddPath(net.Prefix, *route.Path) error {
	netlink.RouteAdd(&netlink.Route{
		Table: n.table,
	})

	return nil
}

func (n *netlinkServer) RemovePath(net.Prefix, *route.Path) bool {
	netlink.RouteDel(&netlink.Route{
		Table: n.table,
	})

	return true
}

func (n *netlinkServer) RouteCount() int64 {
	// TODO
	// netlink.RouteSubscribeWithOptions(
	return 0
}

// Not supported for netlink
func (n *netlinkServer) Register(rt.RouteTableClient) {
}

// Not supported for netlink
func (n *netlinkServer) RegisterWithOptions(rt.RouteTableClient, rt.ClientOptions) {
}

// Not supported for netlink
func (n *netlinkServer) Unregister(rt.RouteTableClient) {
}

// Not needed, since no clients are possible in netlink
func (n *netlinkServer) UpdateNewClient(rt.RouteTableClient) error {
	return nil
}
