package device

import "fmt"

type osAdapter struct {
}

func newOSAdapter() *osAdapter {
	return nil
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
