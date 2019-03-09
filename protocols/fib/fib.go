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
	addPath(pfx bnet.Prefix) error
	removePath(pfx bnet.Prefix, path route.FIBPath) error
	start() error
}

// FIB is forwarding information base
type FIB struct {
	vrf           *vrf.VRF
	osAdapter     fibOsAdapter
	pathTable     map[bnet.Prefix][]route.FIBPath
	paths         []route.FIBPath
	pathsMu       sync.RWMutex
	clientManager *routingtable.ClientManager
}

// New creates a new Netlink object and returns the pointer to it
func New(vrf *vrf.VRF) *FIB {
	n := &FIB{
		vrf:       vrf,
		paths:     make([]route.FIBPath, 0),
		pathTable: make(map[bnet.Prefix][]route.FIBPath),
	}

	n.loadFIB()

	return n
}

// Start the Netlink module
func (f *FIB) Start() error {
	err := f.osAdapter.start()
	if err != nil {
		return errors.Wrap(err, "Unable to start os specific FIB")
	}
	return nil

	// connect all RIBs
	options := routingtable.ClientOptions{
		BestOnly: false,
		EcmpOnly: false,
		MaxPaths: ^uint(0), // max int
	}

	// TODO!!!!!!
	rib, found := f.vrf.RIBByName("inet.0") // v4
	// rib, found := f.vrf.RIBByName("inet6.0") //v6?
	if found {
		// from locRib to FIB
		rib.RegisterWithOptions(f, options)

		// from FIB to locRIB
		f.RegisterWithOptions(rib, options)
	}

	go f.osAdapter.start()

	return nil
}

// RouteCount returns the current count of paths in the FIB
func (f *FIB) RouteCount() int64 {
	f.pathsMu.RLock()
	defer f.pathsMu.RUnlock()
	return int64(len(f.pathTable))
	//return int64(len(f.paths))
}

// AddPath adds the element from the FIB List
func (f *FIB) AddPath(pfx bnet.Prefix, path *route.Path) error {
	var addPath route.FIBPath

	switch path.Type {
	case route.BGPPathType:
		addPath = *route.NewFIBPathFromBgpPath(path.BGPPath)
	case route.FIBPathType:
		addPath = *path.FIBPath
	default:
		return fmt.Errorf("PathType %d is (currently) not supported", path.Type)
	}

	f.pathsMu.Lock()
	defer f.pathsMu.Unlock()

	existingPaths, found := f.pathTable[pfx]
	if !found {
		f.pathTable[pfx] = []route.FIBPath{addPath}

		err := f.osAdapter.addPath(pfx)
		if err != nil {
			err = errors.Wrap(err, "Can't add Path to underlying OS Layer")
			// be transaction safe!
			if !f.cleanupPathAfterTryToAdd(pfx, addPath) {
				err = errors.Wrap(err, "Can't rollback pseudotransaction. Somethings terrible wrong!")
			}
			return err
		}
	}

	if !addPath.ContainedIn(existingPaths) {
		f.pathTable[pfx] = append(existingPaths, addPath)

		err := f.osAdapter.addPath(pfx)
		if err != nil {
			err = errors.Wrap(err, "Can't add Path to underlying OS Layer")
			// be transaction safe!
			if !f.cleanupPathAfterTryToAdd(pfx, addPath) {
				err = errors.Wrap(err, "Can't rollback pseudotransaction. Somethings terrible wrong!")
			}
			return err
		}
	}

	return nil
}

// if something goes wrong during adding the path, clean it up!
func (f *FIB) cleanupPathAfterTryToAdd(pfx bnet.Prefix, path route.FIBPath) bool {
	existingPaths, found := f.pathTable[pfx]
	if !found {
		return false // whooooot???
	}

	for idx, p := range existingPaths {
		if !p.Equals(&path) {
			continue
		}

		err := f.osAdapter.removePath(pfx, path)
		if err != nil {
			log.Errorf("Can't remove Path from underlying OS Layer: %v", err)
		}
		existingPaths = append(existingPaths[:idx], existingPaths[idx+1:]...)
	}
	return true
}

// RemovePath adds the element from the FIB List
func (f *FIB) RemovePath(pfx bnet.Prefix, path *route.Path) bool {
	var delPath route.FIBPath

	switch path.Type {
	case route.BGPPathType:
		delPath = *route.NewFIBPathFromBgpPath(path.BGPPath)
	case route.FIBPathType:
		delPath = *path.FIBPath
	default:
		log.Errorf("PathType %d is (currently) not supported", path.Type)
	}

	f.pathsMu.Lock()
	defer f.pathsMu.Unlock()

	existingPaths, found := f.pathTable[pfx]
	if !found {
		return false
	}

	for idx, p := range existingPaths {
		if !p.Equals(&delPath) {
			continue
		}

		err := f.osAdapter.removePath(pfx, delPath)
		if err != nil {
			log.Errorf("Can't remove Path from underlying OS Layer: %v", err)
		}
		existingPaths = append(existingPaths[:idx], existingPaths[idx+1:]...)
	}

	return found
}

// this function does not aquire a mutex lock! Be careful!
func (f *FIB) addPath(pfx bnet.Prefix, addPaths []route.FIBPath) {
	f.pathsMu.Lock()
	defer f.pathsMu.Unlock()

	existingPaths, found := f.pathTable[pfx]
	if !found {
		f.pathTable[pfx] = addPaths
		return
	}

	for _, pToAdd := range addPaths {
		if !pToAdd.ContainedIn(existingPaths) {
			existingPaths = append(existingPaths, pToAdd)
		}
	}
}

// this function does not aquire a mutex lock! Be careful!
func (f *FIB) removePath(pfx bnet.Prefix, delPaths []route.FIBPath) bool {
	f.pathsMu.Lock()
	defer f.pathsMu.Unlock()

	existingPaths, found := f.pathTable[pfx]
	if !found {
		return false
	}

	for _, pToDel := range delPaths {
		for i, exPath := range existingPaths {
			if !exPath.Equals(&pToDel) {
				continue
			}

			existingPaths = append(f.paths[:i], f.paths[i+1:]...)
		}
	}

	f.pathTable[pfx] = existingPaths
	return true
}

// If inFibButNotIncmpTo=true the diff will show which parts of cmpTo are inside fib,
// if inFibButNotIncmpTo=false the diff will show which parts of cmpTo are not inside the fib
// this function does not aquire a mutex lock! Be careful!
func (f *FIB) compareFibPfxPath(cmpTo []route.PrefixPathsPair, inFibButNotIncmpTo bool) []route.PrefixPathsPair {
	f.pathsMu.Lock()
	defer f.pathsMu.Unlock()

	pfxPathsDiff := make([]route.PrefixPathsPair, 0)

	for _, pfxPath := range cmpTo {
		paths, found := f.pathTable[pfxPath.Pfx]
		if found {
			if inFibButNotIncmpTo {
				pfxPath.Paths = route.FIBPathsDiff(paths, pfxPath.Paths)
			} else {
				pfxPath.Paths = route.FIBPathsDiff(pfxPath.Paths, paths)
			}
		}
		pfxPathsDiff = append(pfxPathsDiff, pfxPath)
	}

	return pfxPathsDiff
}

// Stop stops the device server
func (f *FIB) Stop() {
}

// UpdateNewClient Not supported for NetlinkWriter, since the writer is not observable
func (f *FIB) UpdateNewClient(routingtable.RouteTableClient) error {
	return fmt.Errorf("Not supported")
}

// Register Not supported for NetlinkWriter, since the writer is not observable
func (f *FIB) Register(routingtable.RouteTableClient) {
	log.Panic("Not supported")
}

// RegisterWithOptions Not supported, since the writer is not observable
func (f *FIB) RegisterWithOptions(routingtable.RouteTableClient, routingtable.ClientOptions) {
	log.Panic("Not supported")
}

// Unregister is not supported, since the writer is not observable
func (f *FIB) Unregister(routingtable.RouteTableClient) {
	log.Panic("Not supported")
}

// ClientCount is currently not implemented
func (f *FIB) ClientCount() uint64 {
	return 0
}

// Dump is currently not supported
func (f *FIB) Dump() []*route.Route {
	return nil
}
