package filter

import (
	"testing"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/stretchr/testify/assert"
)

func TestNewAcceptAllFilter(t *testing.T) {
	f := NewAcceptAllFilter()

	_, reject := f.ProcessTerms(net.NewPfx(0, 0), &route.Path{})
	assert.Equal(t, false, reject)
}

func TestNewDrainFilter(t *testing.T) {
	f := NewDrainFilter()

	_, reject := f.ProcessTerms(net.NewPfx(0, 0), &route.Path{})
	assert.Equal(t, true, reject)
}
