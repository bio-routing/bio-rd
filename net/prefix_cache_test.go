package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixCache(t *testing.T) {
	a := &Prefix{
		addr: IP{
			higher:   100,
			lower:    200,
			isLegacy: false,
		},
		len: 64,
	}
	b := &Prefix{
		addr: IP{
			higher:   100,
			lower:    200,
			isLegacy: false,
		},
		len: 64,
	}

	x := a.Dedup()
	y := b.Dedup()

	assert.Equal(t, true, x == y)
}
