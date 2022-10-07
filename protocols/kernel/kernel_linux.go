package kernel

import (
	"fmt"

	"github.com/bio-routing/bio-rd/route"
	"github.com/vishvananda/netlink"

	bnet "github.com/bio-routing/bio-rd/net"
)

const (
	protoBio = 45
)

func (k *Kernel) init() error {
	lk, err := newLinuxKernel()
	if err != nil {
		return fmt.Errorf("unable to initialize linux kernel: %w", err)
	}

	err = lk.init()
	if err != nil {
		return fmt.Errorf("init failed: %w", err)
	}
	k.osKernel = lk
	return nil
}

type linuxKernel struct {
	h      *netlink.Handle
	routes map[*bnet.Prefix]struct{}
}

func newLinuxKernel() (*linuxKernel, error) {
	h, err := netlink.NewHandle()
	if err != nil {
		return nil, fmt.Errorf("unable to get Netlink handle: %w", err)
	}

	return &linuxKernel{
		h:      h,
		routes: make(map[*bnet.Prefix]struct{}),
	}, nil
}

func (lk *linuxKernel) init() error {
	err := lk.cleanup()
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	return nil
}

func (lk *linuxKernel) uninit() error {
	return lk.cleanup()
}

func (lk *linuxKernel) cleanup() error {
	filter := &netlink.Route{
		Protocol: protoBio,
	}

	routes, err := lk.h.RouteListFiltered(0, filter, netlink.RT_FILTER_PROTOCOL)
	if err != nil {
		return fmt.Errorf("unable to get routes: %w", err)
	}

	for i := range routes {
		err = lk.h.RouteDel(&routes[i])
		if err != nil {
			return fmt.Errorf("unable to remove route: %w", err)
		}
	}

	return nil
}

func (lk *linuxKernel) AddPath(pfx *bnet.Prefix, path *route.Path) error {
	r := &netlink.Route{
		Protocol: protoBio,
		Dst:      pfx.GetIPNet(),
		Gw:       path.NextHop().ToNetIP(),
	}

	if _, found := lk.routes[pfx]; !found {
		err := lk.h.RouteAdd(r)
		if err != nil {
			return fmt.Errorf("unable to add route: %w", err)
		}

		lk.routes[pfx] = struct{}{}
		return nil
	}

	err := lk.h.RouteReplace(r)
	if err != nil {
		return fmt.Errorf("unable to replace route: %w", err)
	}

	return nil
}

func (lk *linuxKernel) RemovePath(pfx *bnet.Prefix, path *route.Path) bool {
	if _, found := lk.routes[pfx]; !found {
		return false
	}

	r := &netlink.Route{
		Protocol: protoBio,
		Dst:      pfx.GetIPNet(),
		Gw:       path.NextHop().ToNetIP(),
	}

	err := lk.h.RouteDel(r)
	if err != nil {
		return false
	}

	delete(lk.routes, pfx)
	return true
}
