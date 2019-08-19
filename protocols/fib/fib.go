package fib

import (
	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
)

// FIB is forwarding information base
type FIB struct {
	locRib *locRIB.LocRIB

	//writer *NetlinkWriter
	//reader *NetlinkReader
}

// NewFIB creates a new Netlink object and returns the pointer to it
func NewFIB(options *config.Netlink, locRib *locRIB.LocRIB) *FIB {
	return &FIB{
		locRib: locRib,
		//writer: NewNetlinkWriter(options),
		//reader: NewNetlinkReader(options),
	}
}

// Start the Netlink module
func (f *FIB) Start() {
	// connect all RIBs
	/*options := routingtable.ClientOptions{
		BestOnly: false,
		EcmpOnly: false,
		MaxPaths: ^uint(0), // max int
	}*/

	// 1. from locRib to Kernel
	//f.locRib.RegisterWithOptions(n.writer, options)

	// 2. from Kernel to locRib
	//f.reader.clientManager.RegisterWithOptions(n.locRib, options)

	// Listen for new routes from kernel
	//go n.reader.Read()
}
