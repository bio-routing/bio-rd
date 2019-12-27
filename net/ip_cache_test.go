package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPCache(t *testing.T) {
	a := &IP{
		higher:   100,
		lower:    200,
		isLegacy: false,
	}
	b := &IP{
		higher:   100,
		lower:    200,
		isLegacy: false,
	}

	x := a.Dedup()
	y := b.Dedup()

	assert.Equal(t, true, x == y)
}
