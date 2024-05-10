package rismirror

import (
	"fmt"
	"net"
	"sync"

	"github.com/bio-routing/bio-rd/risclient"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/bio-routing/bio-rd/util/log"
	"google.golang.org/grpc"

	"github.com/bio-routing/bio-rd/cmd/ris/api"
)

// Router represents a router
type Router struct {
	name        string
	address     net.IP
	vrfs        map[uint64]*vrfWithMergedLocRIBs // this is the authoritative data store for VRFs
	vrfsMu      sync.RWMutex
	vrfRegistry *vrf.VRFRegistry // this is only there so that the metrics functionality of the vrf package can be used
}

func newRouter(name string, address net.IP) *Router {
	return &Router{
		name:        name,
		address:     address,
		vrfs:        make(map[uint64]*vrfWithMergedLocRIBs),
		vrfRegistry: vrf.NewVRFRegistry(),
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
	r.vrfsMu.RLock()
	defer r.vrfsMu.RUnlock()

	return r._getVRF(rd)
}

// _getVRF gets a VRF by its RD.
// _getVRF may only be called with the mutex r.vrfsMu acquired
func (r *Router) _getVRF(rd uint64) *vrf.VRF {
	_vrf := r.vrfs[rd]
	if _vrf == nil {
		return nil
	}

	return _vrf.vrf
}

// GetVRFs gets all VRFs
func (r *Router) GetVRFs() []*vrf.VRF {
	r.vrfsMu.RLock()
	defer r.vrfsMu.RUnlock()

	ret := make([]*vrf.VRF, 0, len(r.vrfs))
	for _, v := range r.vrfs {
		ret = append(ret, v.vrf)
	}

	return ret
}

func (r *Router) addVRF(rd uint64, sources []*grpc.ClientConn) {
	r.vrfsMu.Lock()
	defer r.vrfsMu.Unlock()

	v := r.vrfRegistry.CreateVRFIfNotExists(vrf.RouteDistinguisherHumanReadable(rd), rd)
	r.vrfs[rd] = newVRFWithMergedLocRIBs(v.IPv4UnicastRIB(), v.IPv6UnicastRIB(), v)

	for _, src := range sources {
		rc4 := r._connectVRF(rd, src, 4)
		rc6 := r._connectVRF(rd, src, 6)
		// we cannot call v directly here, because the ris clients are only in the specific implementation of the vrf struct.
		r.vrfs[rd].Clients = append(r.vrfs[rd].Clients, rc4, rc6)
	}
}

func (r *Router) removeVRF(rd uint64) error {
	r.vrfsMu.Lock()
	defer r.vrfsMu.Unlock()

	if r._getVRF(rd) == nil {
		return fmt.Errorf("VRF %v in Router %sv does not exist, cannot remove cleanly", rd, r.address.String())
	}

	// for removal, first stop all vrfs
	r._stopVRF(rd)
	// then we clean the map entries
	vrf := r.vrfs[rd]
	r.vrfRegistry.UnregisterVRF(vrf.vrf)
	delete(r.vrfs, rd)
	return nil
}

func (r *Router) dropAllVRFs() {
	for _, vrf := range r.GetVRFs() {
		r.removeVRF(vrf.RD())
	}
}

func (r *Router) _stopVRF(rd uint64) {
	// check if we even have an vrf for this rd
	if _, exists := r.vrfs[rd]; !exists {
		log.Errorf("Trying to stop invalid vrf %v", rd)
		return
	}
	v := r.vrfs[rd]
	// first step: stop all ris clients
	for _, rc := range v.Clients {
		rc.Stop()
	}
	// now we can clear the vrfClient list
	r.vrfs[rd].Clients = make([]*risclient.RISClient, 0)
}

func (r *Router) _connectVRF(rd uint64, src *grpc.ClientConn, afi uint8) *risclient.RISClient {
	rc := risclient.New(&risclient.Request{
		Router:          r.name,
		VRFRD:           rd,
		AFI:             apiAFI(afi),
		AllowUnreadyRib: true,
	}, src, r.vrfs[rd].getRIB(afi))

	rc.Start()
	return rc
}

func apiAFI(afi uint8) api.ObserveRIBRequest_AFISAFI {
	if afi == 6 {
		return api.ObserveRIBRequest_IPv6Unicast
	}

	return api.ObserveRIBRequest_IPv4Unicast
}
