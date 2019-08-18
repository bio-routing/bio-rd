package filter

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestNewAcceptAllFilter(t *testing.T) {
	f := NewAcceptAllFilter()

	res := f.Process(net.NewPfx(net.IPv4(0), 0), &route.Path{})
	assert.Equal(t, false, res.Reject)
}

func TestNewDrainFilter(t *testing.T) {
	f := NewDrainFilter()

	res := f.Process(net.NewPfx(net.IPv4(0), 0), &route.Path{})
	assert.Equal(t, true, res.Reject)
}
