package vamp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVRFTracker(t *testing.T) {
	vt := newVRFTracker()

	_, err := vt.getVRFID("main")
	if err == nil {
		t.Fatalf("Error expected when querying non-existing VRF, but didn't get one")
	}

	mainID, err := vt.registerVRF("main")
	assert.NoError(t, err, "Got error while registering VRF main")
	assert.Equal(t, uint32(0), mainID, "VRF main")

	id, err := vt.getVRFID("main")
	if err != nil {
		t.Fatalf("Got unexpected error while querying existing VRF main")
	}
	assert.Equal(t, uint32(0), id, "VRF main")

	_, err = vt.registerVRF("main")
	assert.Error(t, err, "Got no error while re-registering VRF main")

	fourtytwoId, err := vt.registerVRF("fourtytwo")
	assert.NoError(t, err, "Got error while registering VRF fourtytwo")
	assert.Equal(t, uint32(1), fourtytwoId, "VRF fourtytwo")

	id, err = vt.getVRFID("fourtytwo")
	if err != nil {
		t.Fatalf("Got unexpected error while querying existing VRF fourtytwo")
	}
	assert.Equal(t, uint32(1), id, "VRF fourtytwo")

	vt.unregisterVRF("main")
	assert.NoError(t, err, "Got error while unregistering VRF main")

	_, err = vt.getVRFID("main")
	if err == nil {
		t.Fatalf("Error expected when querying unregisted VRF main, but didn't get one")
	}
}
