package config

import (
	"fmt"
	"strconv"
	"strings"
)

type RoutingInstance struct {
	// description: |
	//   Name of the routing instance
	Name                       string `yaml:"name"`
	// description: |
	//   String to be used as a route distinguisher.
	//   The format has to be <uint32>:<uint32>. Using IP addresses is *not* allowed
	RouteDistinguisher         string `yaml:"route_distinguisher"`
	// docgen:nodoc
	InternalRouteDistinguisher uint64
	// description: |
	//   Routing options for this routing instance. See main config documentation for details
	RoutingOptions             *RoutingOptions `yaml:"routing_options"`
	// description: |
	//   Protocols for this routing instance. See the main protocols documentation for details
	Protocols                  *Protocols `yaml:"protocols"`
}

func (ri *RoutingInstance) load() error {
	err := ri.loadRD()
	if err != nil {
		return fmt.Errorf("unable to load route distinguisher: %w", err)
	}

	return nil
}

func (ri *RoutingInstance) loadRD() error {
	parts := strings.Split(ri.RouteDistinguisher, ":")
	if len(parts) != 2 {
		return fmt.Errorf("Invalid format: %q", ri.RouteDistinguisher)
	}

	a, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("Invalid format: %q", ri.RouteDistinguisher)
	}

	b, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("Invalid format: %q", ri.RouteDistinguisher)
	}

	rd := uint64(b)
	rd += uint64(a) << 32

	ri.InternalRouteDistinguisher = rd
	return nil
}
