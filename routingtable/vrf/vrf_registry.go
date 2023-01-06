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
	vrfs map[string]*VRF
	mu   sync.RWMutex
}

func NewVRFRegistry() *VRFRegistry {
	return &VRFRegistry{
		vrfs: make(map[string]*VRF),
	}
}

func (r *VRFRegistry) CreateVRFIfNotExists(name string, rd uint64) *VRF {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.vrfs[name]; ok {
		return r.vrfs[name]
	}

	r.vrfs[name] = newUntrackedVRF(name, rd)
	r.vrfs[name].CreateIPv4UnicastLocRIB("inet.0")
	r.vrfs[name].CreateIPv6UnicastLocRIB("inet6.0")

	return r.vrfs[name]
}

// registerVRF adds the given VRF from the global registry.
// An error is returned if there is already a VRF registered with the same route distinguisher.
func (r *VRFRegistry) registerVRF(v *VRF) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.vrfs[v.name]; ok {
		return fmt.Errorf("a VRF with the rd '%d' already exists", v.routeDistinguisher)
	}

	r.vrfs[v.name] = v
	return nil
}

// unregisterVRF removes the given VRF from the registry.
func (r *VRFRegistry) UnregisterVRF(v *VRF) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.vrfs, v.name)
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
	r.mu.RLock()
	defer r.mu.RUnlock()

	l := make([]*VRF, len(r.vrfs))
	i := 0
	for _, v := range r.vrfs {
		l[i] = v
		i++
	}

	return l
}

// GetGlobalRegistry gets the global registry
func GetGlobalRegistry() *VRFRegistry {
	return globalRegistry
}

// GetVRFByRD gets a VRF by route distinguisher
/*func (r *VRFRegistry) GetVRFByRD(rd uint64) *VRF {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.vrfs[rd]; ok {
		return r.vrfs[rd]
	}

	return nil
}*/

// GetVRFByRD gets a VRF by route distinguisher
func (r *VRFRegistry) GetVRFByName(name string) *VRF {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.vrfs[name]; ok {
		return r.vrfs[name]
	}

	return nil
}
