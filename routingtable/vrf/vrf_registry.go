package vrf

import (
	"fmt"
	"sync"
)

var globalRegistry *vrfRegistry

func init() {
	globalRegistry = &vrfRegistry{
		vrfs: make(map[string]*VRF),
	}
}

// vrfRegistry holds a reference to all active VRFs. Every VRF have to have a different name.
type vrfRegistry struct {
	vrfs map[string]*VRF
	mu   sync.Mutex
}

// registerVRF adds the given VRF from the global registry.
// An error is returned if there is already a VRF registered with the same name.
func (r *vrfRegistry) registerVRF(v *VRF) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, found := r.vrfs[v.name]
	if found {
		return fmt.Errorf("a VRF with the name '%s' already exists", v.name)
	}

	r.vrfs[v.name] = v
	return nil
}

// unregisterVRF removes the given VRF from the global registry.
func (r *vrfRegistry) unregisterVRF(v *VRF) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.vrfs, v.name)
}
