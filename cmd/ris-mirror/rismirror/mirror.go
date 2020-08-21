package rismirror

import (
	"fmt"
	"net"
	"sync"

	"github.com/bio-routing/bio-rd/cmd/ris-mirror/rtmirror"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
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

func (rism *RISMirror) AddTarget(rtrName string, address net.IP, vrf uint64, sources []*grpc.ClientConn) {
	rism.routersMu.Lock()
	defer rism.routersMu.Unlock()

	if _, exists := rism.routers[rtrName]; !exists {
		rism.routers[rtrName] = newRouter(rtrName, address)
	}

	v := rism.routers[rtrName].(*Router).vrfRegistry.CreateVRFIfNotExists(fmt.Sprintf("%d", vrf), vrf)

	rtmirror.New(sources, rtmirror.Config{
		Router: rtrName,
		VRF:    v,
	})
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
