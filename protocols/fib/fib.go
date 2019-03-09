package fib

import (
	"sync"

	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/pkg/errors"
)

type fibOsAdapter interface {
	addPath(path route.FIBPath) error
	removePath(path route.FIBPath) error
	pathCount() int64
	start() error
}

// FIB is forwarding information base
type FIB struct {
	vrf       *vrf.VRF
	osAdapter fibOsAdapter
	pathsMu   sync.RWMutex
	paths     []route.FIBPath
	done      chan struct{}
}

// New creates a new Netlink object and returns the pointer to it
func New(vrf *vrf.VRF) *FIB {
	n := &FIB{
		vrf:   vrf,
		paths: make([]route.FIBPath, 0),
		done:  make(chan struct{}),
	}

	n.loadFIB()

	return n
}

// Start the Netlink module
func (f *FIB) Start() error {
	err := f.osAdapter.start()
	if err != nil {
		return errors.Wrap(err, "Unable to start os specific FIB")
	}
	return nil

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

// Stop stops the device server
func (f *FIB) Stop() {
	close(f.done)
}
