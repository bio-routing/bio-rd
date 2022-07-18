package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestExtendedISReachabilityTLVSerialize(t *testing.T) {
	tests := []struct {
		name     string
		tlv      *ExtendedISReachabilityTLV
		expected []byte
	}{
		{
			name: "Test #1",
			tlv: &ExtendedISReachabilityTLV{
				TLVType:   22,
				TLVLength: 21,
				Neighbors: []*ExtendedISReachabilityNeighbor{
					{
						NeighborID: types.SourceID{
							SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
							CircuitID: 100,
						},
						Metric:       123,
						SubTLVLength: 8,
						SubTLVs: []TLV{
							NewLinkLocalRemoteIdentifiersSubTLV(1000, 2000),
						},
					},
				},
			},
			expected: []byte{
				22,
				21,
				10, 20, 30, 40, 50, 60,
				100,
				0, 0, 123,
				8,
				4,
				8,
				0, 0, 0x3, 0xe8,
				0, 0, 0x7, 0xd0,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.tlv.Serialize(buf)
		assert.Equal(t, test.expected, buf.Bytes(), test.name)
	}
}

func TestExtendedISReachabilityNeighborAddSubTLV(t *testing.T) {
	tests := []struct {
		name     string
		neighbor *ExtendedISReachabilityNeighbor
		addTLV   TLV
		expected *ExtendedISReachabilityNeighbor
	}{
		{
			name: "Test #1",
			neighbor: NewExtendedISReachabilityNeighbor(types.NewSourceID(
				types.SystemID{1, 2, 3, 4, 5, 6},
				0,
			), 16777215),
			addTLV: &IPv4AddressSubTLV{
				TLVType:   6,
				TLVLength: 4,
				Address:   111,
			},
			expected: &ExtendedISReachabilityNeighbor{
				NeighborID: types.NewSourceID(
					types.SystemID{1, 2, 3, 4, 5, 6},
					0,
				),
				Metric:       16777215,
				SubTLVLength: 6,
				SubTLVs: []TLV{
					&IPv4AddressSubTLV{
						TLVType:   6,
						TLVLength: 4,
						Address:   111,
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.neighbor.AddSubTLV(test.addTLV)
		assert.Equal(t, test.expected, test.neighbor, test.name)
	}
}

func TestIPv4AddressSubTLVSerialize(t *testing.T) {
	tests := []struct {
		name     string
		tlv      *IPv4AddressSubTLV
		expected []byte
	}{
		{
			name: "Test #1",
			tlv: &IPv4AddressSubTLV{
				TLVType:   IPv4InterfaceAddressSubTLVType,
				TLVLength: 4,
				Address:   111,
			},
			expected: []byte{6, 4, 0, 0, 0, 111},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.tlv.Serialize(buf)
		assert.Equal(t, test.expected, buf.Bytes(), test.name)
	}
}

func TestNewIPv4InterfaceAddressSubTLV(t *testing.T) {
	tlv := NewIPv4InterfaceAddressSubTLV(111)
	expectred := &IPv4AddressSubTLV{
		TLVType:   6,
		TLVLength: 4,
		Address:   111,
	}

	assert.Equal(t, expectred, tlv)
}

func TestNewIPv4NeighborAddressSubTLV(t *testing.T) {
	tlv := NewIPv4NeighborAddressSubTLV(111)
	expectred := &IPv4AddressSubTLV{
		TLVType:   8,
		TLVLength: 4,
		Address:   111,
	}

	assert.Equal(t, expectred, tlv)
}
