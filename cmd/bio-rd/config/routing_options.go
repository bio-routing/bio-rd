package config

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
)

type RoutingOptions struct {
	// description: |
	//   List of static routes to install in the RIB
	//   Still not implemented
	StaticRoutes []StaticRoute `yaml:"static_routes"`
	// description: |
	//   32-bit number to serve as router id. Must have the format x.x.x.x
	RouterID string `yaml:"router_id"`
	// docgen:nodoc
	RouterIDUint32 uint32
	// description: |
	//   32-bit autonomous system number
	AutonomousSystem uint32 `yaml:"autonomous_system"`
}

func (r *RoutingOptions) load() error {
	addr, err := bnet.IPFromString(r.RouterID)
	if err != nil {
		return fmt.Errorf("unable to parse router id: %w", err)
	}
	r.RouterIDUint32 = uint32(addr.Lower())

	return nil
}
