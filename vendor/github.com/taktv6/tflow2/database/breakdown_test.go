package database

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBreakdownKeyString(t *testing.T) {
	assert := assert.New(t)

	// Empty Key
	key := BreakdownKey{}
	assert.Equal("", key.Join("%s:%s"))

	// Set one key
	key.set("DstPort", "23")
	assert.Equal(key.get("DstPort"), "23")
	assert.Equal("DstPort:23", key.Join("%s:%s"))

	// Set all keys
	for i := range breakdownLabels {
		key[i] = strconv.Itoa(i)
	}
	assert.Equal("Family:2,SrcAddr:3,DstAddr:4,Protocol:5,IntIn:6,IntOut:7,NextHop:8,SrcAsn:9,DstAsn:10,NextHopAsn:11,SrcPfx:12,DstPfx:13,SrcPort:14,DstPort:15,IntInName:16,IntOutName:17", key.Join("%s:%s"))
}

func TestBreakdownFlags(t *testing.T) {
	assert := assert.New(t)

	// Defaults
	key := BreakdownFlags{}
	assert.False(key.DstAddr)

	// Enable all
	assert.NoError(key.Set([]string{"Family", "SrcAddr", "DstAddr", "Protocol", "IntIn", "IntOut", "NextHop", "SrcAsn", "DstAsn", "NextHopAsn", "SrcPfx", "DstPfx", "SrcPort", "DstPort"}))
	assert.True(key.DstAddr)
	assert.Equal(14, key.Count())

	// Invalid key
	assert.EqualError(key.Set([]string{"foobar"}), "invalid breakdown key: foobar")
}

func TestGetBreakdownLabels(t *testing.T) {
	assert := assert.New(t)

	labels := GetBreakdownLabels()
	assert.NotNil(labels)
	assert.Contains(labels, "SrcAddr")
}

// reverse mapping for breakdownLabels
func breakdownIndex(key string) int {
	for i, k := range breakdownLabels {
		if k == key {
			return i
		}
	}
	panic("invalid breakdown label: " + key)
}

// set Sets the value of a field
func (bk *BreakdownKey) set(key string, value string) {
	bk[breakdownIndex(key)] = value
}

// get returns the value of a field
func (bk *BreakdownKey) get(key string) string {
	return bk[breakdownIndex(key)]
}
