package packetv3_test

import (
	"testing"

	ospf "github.com/bio-routing/bio-rd/protocols/ospf/packetv3"
	"github.com/stretchr/testify/assert"
)

func TestLSATypeFlooding(t *testing.T) {
	tests := []struct {
		input              ospf.LSAType
		expectUnknownFlood bool
		expectedFlooding   ospf.FloodingScope
	}{
		{
			input:            ospf.LSATypeRouter,
			expectedFlooding: ospf.FloodArea,
		},
		{
			input:            ospf.LSATypeNetwork,
			expectedFlooding: ospf.FloodArea,
		},
		{
			input:            ospf.LSATypeInterAreaPrefix,
			expectedFlooding: ospf.FloodArea,
		},
		{
			input:            ospf.LSATypeInterAreaRouter,
			expectedFlooding: ospf.FloodArea,
		},
		{
			input:            ospf.LSATypeASExternal,
			expectedFlooding: ospf.FloodAS,
		},
		{
			input:            ospf.LSATypeDeprecated,
			expectedFlooding: ospf.FloodArea,
		},
		{
			input:            ospf.LSATypeNSSA,
			expectedFlooding: ospf.FloodArea,
		},
		{
			input:            ospf.LSATypeLink,
			expectedFlooding: ospf.FloodLinkLocal,
		},
		{
			input:            ospf.LSATypeIntraAreaPrefix,
			expectedFlooding: ospf.FloodArea,
		},
		{
			// Unknown with local scope
			input:              0x0022,
			expectUnknownFlood: false,
		},
		{
			// Unknown with flooding scope
			input:              0xa022,
			expectUnknownFlood: true,
			expectedFlooding:   ospf.FloodArea,
		},
		{
			// Unknown with reserved flooding scope
			input:            0x6022,
			expectedFlooding: ospf.FloodReserved,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expectedFlooding, test.input.FloodingScope())
		assert.Equal(t, test.expectUnknownFlood, test.input.FloodIfUnknown())
	}
}
