package device

import "fmt"

type osAdapter struct {
}

func newOSAdapter(srv *Server) *osAdapter {
	return nil
}

func (o *osAdapter) start() error {
	return fmt.Errorf("Not implemented")
}

func (ds *Server) monitorAddrs() error {
	return fmt.Errorf("Not implemented")
}

func (ds *Server) monitorLinks() error {
	return fmt.Errorf("Not implemented")
}

func (ds *Server) getLinkState(devName string) *LinkUpdate {
	return nil
}
