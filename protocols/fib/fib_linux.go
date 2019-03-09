package fib

import (
	"fmt"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func (f *FIB) loadFIB() {
	f.osAdapter = newOSFIBLinux(f)
}

type osFibAdapterLinux struct {
	fib *FIB
}

func newOSFIBLinux(f *FIB) *osFibAdapterLinux {
	fib := &osFibAdapterLinux{
		fib: f,
	}

	return fib
}

func (f *osFibAdapterLinux) addPath(pfx bnet.Prefix) error {
	route, err := f.createRoute(pfx, f.fib.pathTable[pfx])
	if err != nil {
		return errors.Wrap(err, "Could not create route from prefix and path: %v")
	}

	log.WithFields(log.Fields{
		"Prefix": pfx.String(),
		"Table":  route.Table,
	}).Debug("AddPath to netlink")

	err = netlink.RouteAdd(route)
	if err != nil {
		log.Errorf("Error while adding route: %v", err)
		return errors.Wrap(err, "Error while adding route")
	}

	return nil
}

func (f *osFibAdapterLinux) removePath(pfx bnet.Prefix, path route.FIBPath) error {
	nlRoute, err := f.createRoute(pfx, []route.FIBPath{path})
	if err != nil {
		return errors.Wrap(err, "Could not create route from prefix and path: %v")
	}

	log.WithFields(log.Fields{
		"Prefix": pfx.String(),
	}).Debug("Remove from netlink")

	err = netlink.RouteDel(nlRoute)
	if err != nil {
		return errors.Wrap(err, "Error while removing route")
	}

	return nil
}

// create a route from a prefix and a path
func (f *osFibAdapterLinux) createRoute(pfx bnet.Prefix, paths []route.FIBPath) (*netlink.Route, error) {
	route := &netlink.Route{
		Dst:      pfx.GetIPNet(),
		Table:    int(f.fib.vrf.ID()),
		Protocol: route.ProtoBio,
	}

	multiPath := make([]*netlink.NexthopInfo, 0)

	for _, path := range paths {
		nextHop := &netlink.NexthopInfo{
			Gw: path.NextHop.Bytes(),
		}
		multiPath = append(multiPath, nextHop)
	}

	if len(multiPath) == 1 {
		route.Gw = multiPath[0].Gw
	} else if len(multiPath) > 1 {
		route.MultiPath = multiPath
	} else {
		return nil, fmt.Errorf("No destination address specified. At least one NextHop has to be specified in path")
	}

	return route, nil
}

func (f *osFibAdapterLinux) start() error {
	// Start fetching the kernel routes after the hold timer expires
	// TODO: time.Sleep(nr.options.HoldTime)

	errCnt := 3

	go func() {
		// TODO graceful shutdown this loop in case of closing
		for {
			// Family doesn't matter since I only filter by rt_table
			nlKernelRoutes, err := netlink.RouteListFiltered(bnet.IPFamily4, &netlink.Route{
				Table: int(f.fib.vrf.ID()),
			}, netlink.RT_FILTER_TABLE)

			if err != nil {
				if errCnt--; errCnt == 0 {
					log.WithError(err).Panic("Failed to read routes from kernel")
				} else {
					log.WithError(err).Error("Failed to read routes from kernel")
				}
			}

			fibPfxPaths := convertNlRouteToFIBPath(nlKernelRoutes, true)

			diffContainedInFib := f.fib.compareFibPfxPath(fibPfxPaths, true)
			diffContainedInKernel := f.fib.compareFibPfxPath(fibPfxPaths, false)

			// remove diffContainedInFib from FIB
			for _, staleInFib := range diffContainedInFib {
				f.fib.removePath(staleInFib.Pfx, staleInFib.Paths)
			}

			// add diffContainedInKernel to FIB
			for _, newToFib := range diffContainedInKernel {
				f.fib.addPath(newToFib.Pfx, newToFib.Paths)
			}

			// f.fib.callUpdate(fromKernel)

			// TODO: time.Sleep(nr.options.UpdateInterval)
			time.Sleep(2 * time.Second)
		}
	}()
	return nil
}

// convertNlRouteToFIBPath creates a new route.FIBPath object from a netlink.Route object
func convertNlRouteToFIBPath(netlinkRoutes []netlink.Route, fromKernel bool) []route.PrefixPathsPair {
	fibPfxPaths := make([]route.PrefixPathsPair, 0)

	for _, r := range netlinkRoutes {
		var src bnet.IP
		var dst bnet.Prefix

		if r.Src == nil && r.Dst == nil {
			log.Debugf("Cannot create NlPath, since source and destination are both nil")
			continue
		}

		if r.Src == nil && r.Dst != nil {
			dst = bnet.NewPfxFromIPNet(r.Dst)
			if dst.Addr().IsIPv4() {
				src = bnet.IPv4(0)
			} else {
				src = bnet.IPv6(0, 0)
			}
		}

		if r.Src != nil && r.Dst == nil {
			src, _ = bnet.IPFromBytes(r.Src)
			if src.IsIPv4() {
				dst = bnet.NewPfx(bnet.IPv4(0), 0)
			} else {
				dst = bnet.NewPfx(bnet.IPv6(0, 0), 0)
			}
		}

		if r.Src != nil && r.Dst != nil {
			src, _ = bnet.IPFromBytes(r.Src)
			dst = bnet.NewPfxFromIPNet(r.Dst)
		}

		paths := make([]route.FIBPath, 0)

		// If it's multipath, we have to create serveral paths to the same prefix
		if len(r.MultiPath) > 0 {
			for _, multiPath := range r.MultiPath {
				nextHop, _ := bnet.IPFromBytes(multiPath.Gw)
				paths = append(paths, route.FIBPath{
					Src:      src,
					NextHop:  nextHop,
					Priority: r.Priority,
					Protocol: r.Protocol,
					Type:     r.Type,
					Table:    r.Table,
					Kernel:   fromKernel,
				})
			}
		} else {
			nextHop, _ := bnet.IPFromBytes(r.Gw)
			paths = append(paths, route.FIBPath{
				Src:      src,
				NextHop:  nextHop,
				Priority: r.Priority,
				Protocol: r.Protocol,
				Type:     r.Type,
				Table:    r.Table,
				Kernel:   fromKernel,
			})
		}

		fibPfxPaths = append(fibPfxPaths, route.PrefixPathsPair{
			Pfx:   dst,
			Paths: paths,
			Err:   nil,
		})
	}

	return fibPfxPaths
}
