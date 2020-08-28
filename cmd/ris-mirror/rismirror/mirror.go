package rismirror

import (
	"fmt"
	"net"
	"sync"

	"github.com/bio-routing/bio-rd/cmd/ris-mirror/rtmirror"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/protocols/ris/metrics"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
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

func (rism *RISMirror) AddTarget(rtrName string, address net.IP, vrfRD uint64, sources []*grpc.ClientConn) {
	rism.routersMu.Lock()
	defer rism.routersMu.Unlock()

	if _, exists := rism.routers[rtrName]; !exists {
		rism.routers[rtrName] = newRouter(rtrName, address)
	}

	v := rism.routers[rtrName].(*Router).vrfRegistry.GetVRFByRD(vrfRD)
	if v == nil {
		v = rism.routers[rtrName].(*Router).vrfRegistry.CreateVRFIfNotExists(fmt.Sprintf("%d", vrfRD), vrfRD)
		rtm := rtmirror.New(rtmirror.Config{
			Router: rtrName,
			VRF:    v,
		})

		rism.routers[rtrName].(*Router).rtMirrors[vrfRD] = rtm
	}
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

func (rism *RISMirror) Metrics() *metrics.RISMirrorMetrics {
	res := &metrics.RISMirrorMetrics{
		Routers: make([]*metrics.RISMirrorRouterMetrics, 0),
	}

	rism.routersMu.Lock()
	defer rism.routersMu.Unlock()

	for _, r := range rism.routers {
		rm := &metrics.RISMirrorRouterMetrics{
			Address:    r.Address(),
			SysName:    r.Name(),
			VRFMetrics: vrf.Metrics(r.(*Router).vrfRegistry),
			// TODO: RISUpstreamStatus: Fill In,
		}

		res.Routers = append(res.Routers, rm)
	}

	return res
}
