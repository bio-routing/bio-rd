package rismirror

import (
	"fmt"
	"net"
	"sync"

	"github.com/bio-routing/bio-rd/cmd/ris-mirror/rismirror/metrics"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/bio-routing/bio-rd/util/log"
	"google.golang.org/grpc"
)

type RISMirror struct {
	routers   map[string]server.RouterInterface
	routersMu sync.Mutex
}

// New creates a new RISMirror
func New() *RISMirror {
	return &RISMirror{
		routers: make(map[string]server.RouterInterface),
	}
}

// AddTarget adds a target to the RISMirror
func (rism *RISMirror) AddTarget(rtrName string, address net.IP, vrfRD uint64, sources []*grpc.ClientConn) {
	rism.routersMu.Lock()
	defer rism.routersMu.Unlock()

	if _, exists := rism.routers[rtrName]; !exists {
		rism.routers[rtrName] = newRouter(rtrName, address)
	}

	r := rism.routers[rtrName].(*Router)
	r.addVRF(vrfRD, sources)
}

func (rism *RISMirror) RemoveTarget(rtrName string, vrfRD uint64) error {
	rism.routersMu.Lock()
	defer rism.routersMu.Unlock()

	if _, exists := rism.routers[rtrName]; !exists {
		return fmt.Errorf("Router %s does not exist, cannot remove cleanly", rtrName)
	}
	r := rism.routers[rtrName].(*Router)
	err := r.removeVRF(vrfRD)
	if err != nil {
		return err
	}
	if len(r.GetVRFs()) == 0 {
		delete(rism.routers, rtrName)
	}
	return nil
}

func (rism *RISMirror) RemoveRouter(rtrName string) {
	rism.routersMu.Lock()
	defer rism.routersMu.Unlock()

	if _, exists := rism.routers[rtrName]; !exists {
		log.Errorf("Router %v does not exist, cannot remove cleanly", rtrName)
		return
	}
	r := rism.routers[rtrName].(*Router)
	r.dropAllVRFs()
	delete(rism.routers, rtrName)
}

// GetRouter gets a router
func (rism *RISMirror) GetRouter(rtr string) server.RouterInterface {
	rism.routersMu.Lock()
	defer rism.routersMu.Unlock()

	if _, exists := rism.routers[rtr]; !exists {
		return nil
	}

	return rism.routers[rtr]
}

// GetRouters gets all routers
func (rism *RISMirror) GetRouters() []server.RouterInterface {
	res := make([]server.RouterInterface, 0)

	for _, r := range rism.routers {
		res = append(res, r)
	}

	return res
}

// Metrics gets a RISMirrors metrics
func (rism *RISMirror) Metrics() *metrics.RISMirrorMetrics {
	res := &metrics.RISMirrorMetrics{
		Routers: make([]*metrics.RISMirrorRouterMetrics, 0),
	}

	rism.routersMu.Lock()
	defer rism.routersMu.Unlock()

	for _, r := range rism.routers {
		rm := &metrics.RISMirrorRouterMetrics{
			Address:            r.Address(),
			SysName:            r.Name(),
			VRFMetrics:         vrf.Metrics(r.(*Router).vrfRegistry),
			InternalVRFMetrics: make([]*metrics.InternalVRFMetrics, 0),
		}

		for rd, v := range r.(*Router).vrfs {
			rm.InternalVRFMetrics = append(rm.InternalVRFMetrics, &metrics.InternalVRFMetrics{
				RD:                             rd,
				MergedLocRIBMetricsIPv4Unicast: v.ipv4Unicast.Metrics(),
				MergedLocRIBMetricsIPv6Unicast: v.ipv6Unicast.Metrics(),
			})
		}

		res.Routers = append(res.Routers, rm)
	}

	return res
}
