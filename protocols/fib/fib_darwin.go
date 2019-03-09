package fib

import (
	"fmt"

	"github.com/bio-routing/bio-rd/route"
)

func (f *FIB) loadFIB() {
	f.osAdapter = newOSFIBDarwin(f)
}

type osFibAdapterDarwin struct {
	fib *FIB
}

func newOSFIBLinux(f *FIB) (*osFibAdapterDarwin, error) {
	fib := &osFibAdapterDarwin{
		fib: f,
	}

	return fib, nil
}

func (fib *osFibAdapterDarwin) addPath(path route.FIBPath) error {
	return fmt.Errorf("Not implemented")
}

func (fib *osFibAdapterDarwin) rmovePath(path route.FIBPath) error {
	return fmt.Errorf("Not implemented")
}

func (fib *osFibAdapterDarwin) pathCount() int64 {
	return 0
}

func (fib *osFibAdapterDarwin) start() error {
	return fmt.Errorf("Not implemented")
}
