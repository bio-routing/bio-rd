package vrf

import (
	"fmt"
	"sync"
)

var globalRegistry *VRFRegistry

func init() {
	globalRegistry = NewVRFRegistry()
}

// VRFRegistry holds a reference to all active VRFs. Every VRF have to have a different name.
type VRFRegistry struct {
	vrfs map[uint64]*VRF
	mu   sync.Mutex
}

func NewVRFRegistry() *VRFRegistry {
	return &VRFRegistry{
		vrfs: make(map[uint64]*VRF),
	}
}

func (r *VRFRegistry) CreateVRFIfNotExists(name string, rd uint64) *VRF {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.vrfs[rd]; ok {
		return r.vrfs[rd]
	}

	r.vrfs[rd] = newUntrackedVRF(name, rd)
	r.vrfs[rd].CreateIPv4UnicastLocRIB("inet.0")
	r.vrfs[rd].CreateIPv6UnicastLocRIB("inet6.0")
	return r.vrfs[rd]
}

// registerVRF adds the given VRF from the global registry.
// An error is returned if there is already a VRF registered with the same route distinguisher.
func (r *VRFRegistry) registerVRF(v *VRF) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.vrfs[v.routeDistinguisher]; ok {
		return fmt.Errorf("a VRF with the rd '%d' already exists", v.routeDistinguisher)
	}

	r.vrfs[v.routeDistinguisher] = v
	return nil
}

// unregisterVRF removes the given VRF from the registry.
func (r *VRFRegistry) UnregisterVRF(v *VRF) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.vrfs, v.routeDistinguisher)
}

// DisposeAll dosposes all VRFs
func (r *VRFRegistry) DisposeAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id := range r.vrfs {
		for _, rib := range r.vrfs[id].ribs {
			rib.Dispose()
		}
		delete(r.vrfs, id)
	}
}

func (r *VRFRegistry) List() []*VRF {
	r.mu.Lock()
	defer r.mu.Unlock()

	l := make([]*VRF, len(r.vrfs))
	i := 0
	for _, v := range r.vrfs {
		l[i] = v
		i++
	}

	return l
}

// GetVRFByRD gets a VRF by it's Route Distinguisher
func GetVRFByRD(rd uint64) *VRF {
	return globalRegistry.GetVRFByRD(rd)
}

// GetGlobalRegistry gets the global registry
func GetGlobalRegistry() *VRFRegistry {
	return globalRegistry
}

// GetVRFByRD gets a VRF by route distinguisher
func (r *VRFRegistry) GetVRFByRD(rd uint64) *VRF {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.vrfs[rd]; ok {
		return r.vrfs[rd]
	}

	return nil
}

// GetVRFByRD gets a VRF by route distinguisher
func (r *VRFRegistry) GetVRFByName(name string) *VRF {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, vrf := range r.vrfs {
		if vrf.name == name {
			return vrf
		}
	}

	return nil
}
