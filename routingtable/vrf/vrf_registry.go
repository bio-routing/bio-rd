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

type vrfRegistry struct {
	vrfs map[string]*VRF
	mu   sync.Mutex
}

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
