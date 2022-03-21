package device

import (
	"fmt"
)

func (ds *Server) loadAdapter() error {
	a, err := newOSAdapterDarwin(ds)
	if err != nil {
		return fmt.Errorf("Unable to create OS X adapter: %w", err)
	}

	ds.osAdapter = a
	return nil
}

type osAdapterDarwin struct {
}

func newOSAdapterDarwin(srv *Server) (*osAdapterDarwin, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (o *osAdapterDarwin) start() error {
	return fmt.Errorf("Not implemented")
}

func (o *osAdapterDarwin) init() error {
	return fmt.Errorf("Not implemented")
}
