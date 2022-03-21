package config

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
)

type RoutingOptions struct {
	StaticRoutes     []StaticRoute `yaml:"static_routes"`
	RouterID         string        `yaml:"router_id"`
	RouterIDUint32   uint32
	AutonomousSystem uint32 `yaml:"autonomous_system"`
}

func (r *RoutingOptions) load() error {
	addr, err := bnet.IPFromString(r.RouterID)
	if err != nil {
		return fmt.Errorf("Unable to parse router id: %w", err)
	}
	r.RouterIDUint32 = uint32(addr.Lower())

	return nil
}
