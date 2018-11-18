package protocolnetlink

import (
	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
)

// Netlink is the netlink module which handles the entire NetlinkCommunication
type Netlink struct {
	locRib *locRIB.LocRIB

	writer *NetlinkWriter
	reader *NetlinkReader
}

// NewNetlink creates a new Netlink object and returns the pointer to it
func NewNetlink(options *config.Netlink, locRib *locRIB.LocRIB) *Netlink {

	n := &Netlink{
		locRib: locRib,
		writer: NewNetlinkWriter(options),
		reader: NewNetlinkReader(options),
	}
	return n
}

// Start the Netlink module
func (n *Netlink) Start() {
	// connect all RIBs
	options := routingtable.ClientOptions{
		BestOnly: false,
		EcmpOnly: false,
		MaxPaths: ^uint(0), // max int
	}

	// 1. from locRib to Kernel
	n.locRib.RegisterWithOptions(n.writer, options)

	// 2. from Kernel to locRib
	n.reader.clientManager.RegisterWithOptions(n.locRib, options)

	// Listen for new routes from kernel
	go n.reader.Read()
}
