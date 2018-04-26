package frontend

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taktv6/tflow2/database"
)

func TestFormatBreakdownKey(t *testing.T) {
	assert := assert.New(t)

	// empty key
	assert.Equal("", formatBreakdownKey(&database.BreakdownKey{}))

	// key with two values
	key := &database.BreakdownKey{}
	key[database.FieldFamily] = "4"
	key[database.FieldDstPfx] = "foo"
	assert.Equal(`Family="4",DstPfx="foo"`, formatBreakdownKey(key))
}
