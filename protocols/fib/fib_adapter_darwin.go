package fib

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
)

func (f *FIB) loadFIB() {
	f.osAdapter = newOSFIBDarwin(f)
}

type osFibAdapterDarwin struct {
	fib *FIB
}

func newOSFIBDarwin(f *FIB) (*osFibAdapterDarwin, error) {
	fib := &osFibAdapterDarwin{
		fib: f,
	}

	return fib, nil
}

func (fib *osFibAdapterDarwin) addPath(pfx bnet.Prefix) error {
	return fmt.Errorf("Not implemented")
}

func (fib *osFibAdapterDarwin) removePath(pfx bnet.Prefix, path *route.FIBPath) error {
	return fmt.Errorf("Not implemented")
}

func (fib *osFibAdapterDarwin) start() error {
	return fmt.Errorf("Not implemented")
}
