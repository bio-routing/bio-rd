package vamp

import (
	"fmt"
	"sync"
)

type vrfTracker struct {
	trackerMu      sync.RWMutex
	vrfNameToIDMap map[string]uint32
	nextID         uint32
}

func newVRFTracker() *vrfTracker {
	return &vrfTracker{
		vrfNameToIDMap: make(map[string]uint32),
	}
}

func (vt *vrfTracker) registerVRF(vrf string) (uint32, error) {
	vt.trackerMu.Lock()
	defer vt.trackerMu.Unlock()

	if _, exists := vt.vrfNameToIDMap[vrf]; exists {
		return 0, fmt.Errorf("VRF %q already registed", vrf)
	}

	vrfID := vt.nextID

	vt.vrfNameToIDMap[vrf] = vrfID
	vt.nextID += 1

	return vrfID, nil
}

func (vt *vrfTracker) unregisterVRF(vrf string) error {
	vt.trackerMu.Lock()
	defer vt.trackerMu.Unlock()

	if _, exists := vt.vrfNameToIDMap[vrf]; !exists {
		return fmt.Errorf("VRF %q not registered", vrf)
	}

	delete(vt.vrfNameToIDMap, vrf)

	return nil
}

func (vt *vrfTracker) getVRFID(vrf string) (uint32, error) {
	vt.trackerMu.RLock()
	defer vt.trackerMu.RUnlock()

	vrfID, exists := vt.vrfNameToIDMap[vrf]
	if !exists {
		return 0, fmt.Errorf("vrf %q has not been registered yet", vrf)
	}

	return vrfID, nil
}
