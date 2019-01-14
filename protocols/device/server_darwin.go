package device

import "fmt"

type osAdapterDarwin struct {
}

func newOSAdapterDarwin(srv *Server) (*osAdapterDarwin, error) {
	return nil, nil
}

func (o *osAdapterDarwin) start() error {
	return fmt.Errorf("Not implemented")
}

func (ds *Server) loadAdapter() error {
	return fmt.Errorf("Not implemented")
}
