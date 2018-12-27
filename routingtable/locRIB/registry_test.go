package locRIB

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Register(t *testing.T) {
	r1, err := New("inet.0")
	assert.NotNil(t, r1)
	assert.Nil(t, err)

	r2, err := New("inet.0")
	assert.Nil(t, r2)
	assert.NotNil(t, err)
}
