package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type RoutingInstance struct {
	Name                       string
	RouteDistinguisher         string
	InternalRouteDistinguisher uint64
	RoutingOptions             *RoutingOptions
	Protocols                  *Protocols
}

func (ri *RoutingInstance) load() error {
	err := ri.loadRD()
	if err != nil {
		return errors.Wrap(err, "Unable to load route distinguisher")
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
