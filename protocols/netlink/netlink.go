package proto_netlink

import (
	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
)

type NetlinkServer struct {
	locRib *locRIB.LocRIB

	writer *NetlinkWriter
	reader *NetlinkReader
}

func NewNetlinkServer(options *config.Netlink, locRib *locRIB.LocRIB) *NetlinkServer {

	n := &NetlinkServer{
		locRib: locRib,
		writer: NewNetlinkWriter(options),
		reader: NewNetlinkReader(options),
	}
	return n
}

func (n *NetlinkServer) Start() {
	// connect all RIBs
	options := routingtable.ClientOptions{
		BestOnly: false,
		EcmpOnly: false,
		MaxPaths: ^uint(0), // max int
	}

	// 1. from locRib to Kernel
	n.locRib.ClientManager.RegisterWithOptions(n.writer, options)

	// 2. vom Kernel to locRib
	n.reader.ClientManager.RegisterWithOptions(n.locRib, options)

	// Listn for new routes from kernel
	go n.reader.Read()
}
