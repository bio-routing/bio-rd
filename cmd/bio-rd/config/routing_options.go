package config

import (
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "Unable to parse router id")
	}
	r.RouterIDUint32 = uint32(addr.Lower())

	return nil
}
