package types

import (
	"testing"

	"github.com/bio-routing/bio-rd/route/api"
	"github.com/stretchr/testify/assert"
)

func TestExtendedCommunityFromProtoExtendedCommunity(t *testing.T) {
	input := &api.ExtendedCommunity{
		Type:    128,
		Subtype: 6, // Flow spec traffic-rate
		Value: []byte{
			0x00, 0x00, // 2-Octet AS = 0
			0x00, 0x00, 0x00, 0x00, // Rate shaper = 0
		},
	}

	expected := ExtendedCommunity{
		Type:    128,
		SubType: 6,
		Value: []byte{
			0x00, 0x00, // 2-Octet AS = 0
			0x00, 0x00, 0x00, 0x00, // Rate shaper = 0
		},
	}

	result := ExtendedCommunityFromProtoExtendedCommunity(input)
	assert.Equal(t, expected, result)
}

func TestExtendedCommunityToProto(t *testing.T) {
	input := ExtendedCommunity{
		Type:    128,
		SubType: 6,
		Value: []byte{
			0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
		},
	}

	expected := &api.ExtendedCommunity{
		Type:    128,
		Subtype: 6,
		Value: []byte{
			0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
		},
	}

	result := input.ToProto()
	assert.Equal(t, expected, result)
}