package vrf

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/routingtable/locRIB"
)

const (
	afiIPv4     = 1
	afiIPv6     = 2
	safiUnicast = 1
)

type addressFamily struct {
	afi  uint16
	safi uint8
}

// VRF a list of RIBs for different address families building a routing instance
type VRF struct {
	name     string
	ribs     map[addressFamily]*locRIB.LocRIB
	mu       sync.Mutex
	ribNames map[string]*locRIB.LocRIB
}

// New creates a new VRF. The VRF is registered automatically to the global VRF registry.
func New(name string) (*VRF, error) {
	v := newUntrackedVRF(name)
	v.CreateIPv4UnicastLocRIB("inet.0")
	v.CreateIPv6UnicastLocRIB("inet6.0")

	err := globalRegistry.registerVRF(v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func newUntrackedVRF(name string) *VRF {
	return &VRF{
		name:     name,
		ribs:     make(map[addressFamily]*locRIB.LocRIB),
		ribNames: make(map[string]*locRIB.LocRIB),
	}
}

// CreateLocRIB creates a local RIB with the given name
func (v *VRF) createLocRIB(name string, family addressFamily) (*locRIB.LocRIB, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	_, found := v.ribNames[name]
	if found {
		return nil, fmt.Errorf("a table with the name '%s' already exists in VRF '%s'", name, v.name)
	}

	rib := locRIB.New(name)
	v.ribs[family] = rib
	v.ribNames[name] = rib

	return rib, nil
}

// CreateIPv4UnicastLocRIB creates a LocRIB for the IPv4 unicast address family
func (v *VRF) CreateIPv4UnicastLocRIB(name string) (*locRIB.LocRIB, error) {
	return v.createLocRIB(name, addressFamily{afi: afiIPv4, safi: safiUnicast})
}

// CreateIPv6UnicastLocRIB creates a LocRIB for the IPv6 unicast address family
func (v *VRF) CreateIPv6UnicastLocRIB(name string) (*locRIB.LocRIB, error) {
	return v.createLocRIB(name, addressFamily{afi: afiIPv6, safi: safiUnicast})
}

// IPv4UnicastRIB returns the local RIB for the IPv4 unicast address family
func (v *VRF) IPv4UnicastRIB() *locRIB.LocRIB {
	return v.ribForAddressFamily(addressFamily{afi: afiIPv4, safi: safiUnicast})
}

// IPv6UnicastRIB returns the local RIB for the IPv6 unicast address family
func (v *VRF) IPv6UnicastRIB() *locRIB.LocRIB {
	return v.ribForAddressFamily(addressFamily{afi: afiIPv6, safi: safiUnicast})
}

func (v *VRF) Name() string {
	return v.name
}

// Unregister removes this VRF from the global registry.
func (v *VRF) Unregister() {
	globalRegistry.unregisterVRF(v)
}

func (v *VRF) ribForAddressFamily(family addressFamily) *locRIB.LocRIB {
	v.mu.Lock()
	defer v.mu.Unlock()

	rib, _ := v.ribs[family]

	return rib
}

// RIBByName returns the RIB for a given name. If there is no RIB with this name, found is false
func (v *VRF) RIBByName(name string) (rib *locRIB.LocRIB, found bool) {
	rib, found = v.ribNames[name]
	return rib, found
}
