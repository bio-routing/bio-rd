package device

import (
	"fmt"

	"github.com/pkg/errors"
)

func (ds *Server) loadAdapter() error {
	a, err := newOSAdapterDarwin(ds)
	if err != nil {
		return errors.Wrap(err, "Unable to create OS X adapter")
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
