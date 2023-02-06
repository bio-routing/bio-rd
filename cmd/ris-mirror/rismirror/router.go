package rismirror

import (
	"fmt"
	"net"

	"github.com/bio-routing/bio-rd/risclient"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"google.golang.org/grpc"

	"github.com/bio-routing/bio-rd/cmd/ris/api"
)

// Router represents a router
type Router struct {
	name        string
	address     net.IP
	vrfRegistry *vrf.VRFRegistry
	vrfs        map[uint64]*_vrf
}

func newRouter(name string, address net.IP) *Router {
	return &Router{
		name:        name,
		address:     address,
		vrfRegistry: vrf.NewVRFRegistry(),
		vrfs:        make(map[uint64]*_vrf),
	}
}

// Name gets the routers name
func (r *Router) Name() string {
	return r.name
}

// Address gets a routers address
func (r *Router) Address() net.IP {
	return r.address
}

func (r *Router) Ready(vrf uint64, afi uint16) (bool, error) {
	return true, nil
}

// GetVRF gets a VRF by its RD
func (r *Router) GetVRF(rd uint64) *vrf.VRF {
	return r.vrfRegistry.GetVRFByName(vrf.RouteDistinguisherHumanReadable(rd))
}

// GetVRFs gets all VRFs
func (r *Router) GetVRFs() []*vrf.VRF {
	return r.vrfRegistry.List()
}

func (r *Router) addVRF(rd uint64, sources []*grpc.ClientConn) {
	v := r.vrfRegistry.CreateVRFIfNotExists(fmt.Sprintf("%d", rd), rd)

	r.vrfs[rd] = newVRF(v.IPv4UnicastRIB(), v.IPv6UnicastRIB())

	for _, src := range sources {
		r.connectVRF(rd, src, 4)
		r.connectVRF(rd, src, 6)
	}
}

func (r *Router) connectVRF(rd uint64, src *grpc.ClientConn, afi uint8) {
	rc := risclient.New(&risclient.Request{
		Router:          r.name,
		VRFRD:           rd,
		AFI:             apiAFI(afi),
		AllowUnreadyRib: true,
	}, src, r.vrfs[rd].getRIB(afi))

	rc.Start()
}

func apiAFI(afi uint8) api.ObserveRIBRequest_AFISAFI {
	if afi == 6 {
		return api.ObserveRIBRequest_IPv6Unicast
	}

	return api.ObserveRIBRequest_IPv4Unicast
}
