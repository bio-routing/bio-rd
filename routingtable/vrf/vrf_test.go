package vrf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWithDuplicate(t *testing.T) {
	_, err := NewDefaultVRF()
	assert.Nil(t, err, "no error on first invocation")

	_, err = NewDefaultVRF()
	assert.NotNil(t, err, "ambigious VRF name")
}

func TestIPv4UnicastRIBWith(t *testing.T) {
	v := newUntrackedVRF("master1", uint32(1))
	rib, err := v.CreateIPv4UnicastLocRIB("inet.0")

	assert.Equal(t, rib, v.IPv4UnicastRIB())
	assert.Nil(t, err, "error must be nil")
}

func TestIPv6UnicastRIB(t *testing.T) {
	v := newUntrackedVRF("master2", uint32(2))
	rib, err := v.CreateIPv6UnicastLocRIB("inet6.0")

	assert.Equal(t, rib, v.IPv6UnicastRIB())
	assert.Nil(t, err, "error must be nil")
}

func TestCreateLocRIBTwice(t *testing.T) {
	v := newUntrackedVRF("master3", uint32(3))
	_, err := v.CreateIPv6UnicastLocRIB("inet6.0")
	assert.Nil(t, err, "error must be nil on first invokation")

	_, err = v.CreateIPv6UnicastLocRIB("inet6.0")
	assert.NotNil(t, err, "error must not be nil on second invokation")
}

func TestRIBByName(t *testing.T) {
	v := newUntrackedVRF("master4", uint32(4))
	rib, _ := v.CreateIPv6UnicastLocRIB("inet6.0")
	assert.NotNil(t, rib, "rib must not be nil after creation")

	foundRIB, found := v.RIBByName("inet6.0")
	assert.True(t, found)
	assert.Exactly(t, rib, foundRIB)
}

func TestName(t *testing.T) {
	v := newUntrackedVRF("foo", 5)
	assert.Equal(t, "foo", v.Name())
}
func TestID(t *testing.T) {
	v := newUntrackedVRF("foo", uint32(6))
	assert.Equal(t, uint32(6), v.ID())
}

func TestUnregister(t *testing.T) {
	vrfName := "registeredVRF"
	vrfID := uint32(7)
	v, err := New(vrfName, vrfID)
	assert.Nil(t, err, "error must be nil on first invokation")

	_, err = New(vrfName, vrfID)
	assert.NotNil(t, err, "error must not be nil on second invokation")

	_, found := globalRegistry.vrfsName[vrfName]
	assert.True(t, found, "vrf must be in global registry")

	_, found = globalRegistry.vrfsID[vrfID]
	assert.True(t, found, "vrf must be in global registry")

	v.Unregister()

	_, found = globalRegistry.vrfsName[vrfName]
	assert.False(t, found, "vrf must not be in global registry")

	_, found = globalRegistry.vrfsID[vrfID]
	assert.False(t, found, "vrf must not be in global registry")
}

func TestGetRIBNames(t *testing.T) {
	vrfName := "namedRIBs"
	vrfID := uint32(8)
	v, err := New(vrfName, vrfID)
	assert.Nil(t, err, "error must be nil on first invokation")

	ribNames := v.GetRIBNames()
	expectedRibNames := []string{"inet.8", "inet6.8"}

	assert.EqualValues(t, expectedRibNames, ribNames, "rib names must match")
}
