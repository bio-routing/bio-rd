package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeHash(t *testing.T) {
	p := &BGPPath{
		ASPath:           "123 456",
		BGPIdentifier:    1,
		Communities:      "(123, 456)",
		EBGP:             false,
		LargeCommunities: "(1, 2, 3)",
		LocalPref:        100,
		MED:              1,
		NextHop:          100,
		PathIdentifier:   5,
		Source:           4,
	}

	assert.Equal(t, "24d5b7681ab221b464a2c772e828628482cbfa4d5c6aac7a8285d33ef99b868a", p.ComputeHash())

	p.LocalPref = 150

	assert.NotEqual(t, "24d5b7681ab221b464a2c772e828628482cbfa4d5c6aac7a8285d33ef99b868a", p.ComputeHash())
}
