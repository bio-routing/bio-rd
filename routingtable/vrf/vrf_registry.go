package vrf

import (
	"fmt"
	"sync"
)

var globalRegistry *vrfRegistry

func init() {
	globalRegistry = &vrfRegistry{
		vrfsName: make(map[string]*VRF),
		vrfsID:   make(map[uint32]*VRF),
	}
}

// vrfRegistry holds a reference to all active VRFs. Every VRF have to have a different name.
type vrfRegistry struct {
	vrfsName map[string]*VRF
	mu       sync.Mutex
	vrfsID   map[uint32]*VRF
}

// registerVRF adds the given VRF from the global registry.
// An error is returned if there is already a VRF registered with the same name.
func (r *vrfRegistry) registerVRF(v *VRF) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, found := r.vrfsName[v.name]
	if found {
		return fmt.Errorf("a VRF with the name '%s' already exists", v.name)
	}

	_, found = r.vrfsID[v.id]
	if found {
		return fmt.Errorf("a VRF with the id '%d' already exists", v.id)
	}

	r.vrfsName[v.name] = v
	r.vrfsID[v.id] = v
	return nil
}

// unregisterVRF removes the given VRF from the global registry.
func (r *vrfRegistry) unregisterVRF(v *VRF) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.vrfsName, v.name)
	delete(r.vrfsID, v.id)
}
