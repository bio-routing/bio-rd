package fib

import (
	"fmt"
	"sync"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type fibOsAdapter interface {
	addPath(pfx bnet.Prefix, paths []*route.FIBPath) error
	removePath(pfx bnet.Prefix, path *route.FIBPath) error
}

// FIB is forwarding information base
type FIB struct {
	vrf           *vrf.VRF
	osAdapter     fibOsAdapter
	pathTable     map[bnet.Prefix][]*route.FIBPath
	pathsMu       sync.RWMutex
	clientManager *routingtable.ClientManager
}

// New creates a new Netlink object and returns the pointer to it
func New(v *vrf.VRF) (*FIB, error) {
	if v == nil {
		return nil, fmt.Errorf("Cannot create FIB: No VRF given. Please use at least default VRF")
	}

	n := &FIB{
		vrf:       v,
		pathTable: make(map[bnet.Prefix][]*route.FIBPath),
	}

	n.loadOSAdapter()
	n.clientManager = routingtable.NewClientManager(n)

	return n, nil
}

// Start the Netlink module
func (f *FIB) Start() error {
	if f.osAdapter == nil {
		return fmt.Errorf("osAdapter is not loaded correctly")
	}

	// connect all RIBs
	options := routingtable.ClientOptions{
		BestOnly: false,
		EcmpOnly: false,
		MaxPaths: ^uint(0), // max int
	}

	// register to all ribs in VRF
	vrfRIBs := f.vrf.GetRIBNames()
	for _, ribName := range vrfRIBs {
		rib, found := f.vrf.RIBByName(ribName)
		if !found {
			continue
		}

		// from locRib to FIB
		rib.RegisterWithOptions(f, options)
	}

	return nil
}

// RouteCount returns the current count of paths in the FIB
func (f *FIB) RouteCount() int64 {
	f.pathsMu.RLock()
	defer f.pathsMu.RUnlock()

	fibCount := int64(0)

	for _, paths := range f.pathTable {
		fibCount += int64(len(paths))
	}

	return fibCount
}

// AddPath is called from the RIB when an path is added there
func (f *FIB) AddPath(pfx bnet.Prefix, path *route.Path) error {
	// Convert Path to FIBPath
	addPath, err := route.NewFIBPathFromPath(path)
	if err != nil {
		return errors.Wrap(err, "Could not convert path to FIB path")
	}

	f.pathsMu.Lock()
	defer f.pathsMu.Unlock()

	paths, found := f.pathTable[pfx]
	newPath := false
	if found {
		if !addPath.ContainedIn(paths) {
			paths = append(paths, addPath)
			newPath = true
		}
	} else {
		paths = []*route.FIBPath{addPath}
		newPath = true
	}

	if newPath {
		err = f.osAdapter.addPath(pfx, paths)
		if err != nil {
			return errors.Wrap(err, "Can't add Path to underlying OS Layer")
		}

		// Save new paths
		f.pathTable[pfx] = paths
	}

	return nil
}

// RemovePath adds the element from the FIB List
// returns true if something was removed, false otherwise
func (f *FIB) RemovePath(pfx bnet.Prefix, path *route.Path) bool {
	// Convert Path to FIBPath
	delPath, err := route.NewFIBPathFromPath(path)
	if err != nil {
		log.Errorf("Could not convert path to FIB path: %v", err)
		return false
	}

	f.pathsMu.Lock()
	defer f.pathsMu.Unlock()

	existingPaths, found := f.pathTable[pfx]
	if !found {
		return false
	}

	pathCountBeforeDel := len(existingPaths)

	for idx, p := range existingPaths {
		if !p.Equals(delPath) {
			continue
		}

		err := f.osAdapter.removePath(pfx, delPath)
		if err != nil {
			log.Errorf("Can't remove Path from underlying OS Layer: %v", err)
			// TODO: continue here and let the path in the paths table, or should we just log the error and remove the path regardles if the underlaying operation fails
		}

		// remove path from existing paths
		existingPaths = append(existingPaths[:idx], existingPaths[idx+1:]...)
	}

	// Save
	f.pathTable[pfx] = existingPaths

	return pathCountBeforeDel > len(existingPaths)
}

// If inFibButNotIncmpTo=true the diff will show which parts of cmpTo are inside fib,
// if inFibButNotIncmpTo=false the diff will show which parts of cmpTo are not inside the fib
// this function does not aquire a mutex lock! Be careful!
func (f *FIB) compareFibPfxPath(cmpTo []*route.PrefixPathsPair, inFibButNotIncmpTo bool) []*route.PrefixPathsPair {
	pfxPathsDiff := make([]*route.PrefixPathsPair, 0)

	if len(cmpTo) == 0 {
		if inFibButNotIncmpTo {
			for key, value := range f.pathTable {
				pfxPathsDiff = append(pfxPathsDiff, &route.PrefixPathsPair{
					Pfx:   key,
					Paths: value,
				})
			}
		}

		return pfxPathsDiff
	}

	f.pathsMu.Lock()
	defer f.pathsMu.Unlock()

	if len(f.pathTable) == 0 {
		if inFibButNotIncmpTo {
			return pfxPathsDiff
		}

		return cmpTo
	}

	for _, pfxPathPair := range cmpTo {
		paths, found := f.pathTable[pfxPathPair.Pfx]

		if found {
			if inFibButNotIncmpTo {
				pfxPathPair.Paths = route.FIBPathsDiff(paths, pfxPathPair.Paths)
			} else {
				pfxPathPair.Paths = route.FIBPathsDiff(pfxPathPair.Paths, paths)
			}

			pfxPathsDiff = append(pfxPathsDiff, pfxPathPair)
		}
	}

	return pfxPathsDiff
}

// Stop stops the device server
func (f *FIB) Stop() {
}

// UpdateNewClient Not supported for NetlinkWriter, since the writer is not observable
func (f *FIB) UpdateNewClient(routingtable.RouteTableClient) error {
	f.pathsMu.RLock()
	f.pathsMu.RUnlock()

	for _, client := range f.clientManager.Clients() {
		for pfx, addPaths := range f.pathTable {
			for _, addP := range addPaths {
				client.AddPath(pfx, &route.Path{
					Type:    route.FIBPathType,
					FIBPath: addP,
				})
			}
		}
	}

	return nil
}

// Register Not supported for NetlinkWriter, since the writer is not observable
func (f *FIB) Register(client routingtable.RouteTableClient) {
	f.clientManager.RegisterWithOptions(client, routingtable.ClientOptions{BestOnly: true})
}

// RegisterWithOptions Not supported, since the writer is not observable
func (f *FIB) RegisterWithOptions(client routingtable.RouteTableClient, opt routingtable.ClientOptions) {
	f.clientManager.RegisterWithOptions(client, opt)
}

// Unregister is not supported, since the writer is not observable
func (f *FIB) Unregister(client routingtable.RouteTableClient) {
	f.clientManager.Unregister(client)
}

// ClientCount returns how many clients are connected
func (f *FIB) ClientCount() uint64 {
	return f.clientManager.ClientCount()
}

// Dump is currently not supported
func (f *FIB) Dump() []*route.Route {
	return nil
}
