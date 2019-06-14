package kernel

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"

	bnet "github.com/bio-routing/bio-rd/net"
)

const (
	protoBio = 45
)

func (k *Kernel) init() error {
	lk, err := newLinuxKernel()
	if err != nil {
		return errors.Wrap(err, "Unable to initialize linux kernel")
	}

	err = lk.init()
	if err != nil {
		return errors.Wrap(err, "Init failed")
	}
	k.osKernel = lk
	return nil
}

func (k *Kernel) uninit() error {
	return k.osKernel.uninit()
}

type linuxKernel struct {
	h      *netlink.Handle
	routes map[bnet.Prefix]struct{}
}

func newLinuxKernel() (*linuxKernel, error) {
	h, err := netlink.NewHandle()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get Netlink handle")
	}

	return &linuxKernel{
		h:      h,
		routes: make(map[bnet.Prefix]struct{}),
	}, nil
}

func (lk *linuxKernel) init() error {
	err := lk.cleanup()
	if err != nil {
		return errors.Wrap(err, "Cleanup failed")
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
		return errors.Wrap(err, "Unable to get routes")
	}

	for i := range routes {
		err = lk.h.RouteDel(&routes[i])
		if err != nil {
			return errors.Wrap(err, "Unable to remove route")
		}
	}

	return nil
}

func (lk *linuxKernel) AddPath(pfx net.Prefix, path *route.Path) error {
	r := &netlink.Route{
		Protocol: protoBio,
		Dst:      pfx.GetIPNet(),
		Gw:       path.NextHop().ToNetIP(),
	}

	if _, found := lk.routes[pfx]; !found {
		err := lk.h.RouteAdd(r)
		if err != nil {
			return errors.Wrap(err, "Unable to add route")
		}

		lk.routes[pfx] = struct{}{}
		return nil
	}

	err := lk.h.RouteReplace(r)
	if err != nil {
		return errors.Wrap(err, "Unable to replace route")
	}

	return nil
}

func (lk *linuxKernel) RemovePath(pfx net.Prefix, path *route.Path) bool {
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
