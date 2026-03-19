package packet

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadRouterCapabilityTLV(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		expectErr bool
		expected  *RouterCapabilityTLV
	}{
		{
			name:      "empty input",
			input:     []byte{},
			expectErr: true,
		},
		{
			name:      "valid input without sub TLVs",
			input:     []byte{1, 2, 3, 4, 5},
			expectErr: false,
			expected: &RouterCapabilityTLV{
				TLVType:   242,
				TLVLength: 5,
				RouterID:  0x01020304,
				Flags:     5,
				SubTLVs:   nil,
			},
		},
		{
			name:      "valid input with sub TLVs",
			input:     []byte{1, 2, 3, 4, 5, 255, 1, 3},
			expectErr: false,
			expected: &RouterCapabilityTLV{
				TLVType:   242,
				TLVLength: 8,
				RouterID:  0x01020304,
				Flags:     5,
				SubTLVs: []TLV{
					&UnknownTLV{TLVType: 255, TLVLength: 1, TLVValue: []byte{3}},
				},
			},
		},
		{
			name: "real world input from junos with SR enabled",
			input: []byte{
				0x0a, 0x00, 0x01, 0x00, // Router ID
				0x00,       // Flags
				0x02, 0x09, // SR capability TLV
				0xc0,             // SR capability TLV flags
				0x00, 0x03, 0xe8, // SR capability TLV range
				0x01, 0x03, 0x00, 0x13, 0x88, // SID/Label TLV
				0x13, 0x02, 0x00, 0x01, // SR alrotighms TLV
				0x17, 0x04, 0x01, 0x03, 0x02, 0x10, // Node Maximum SID Depth TLV
				0x0c, 0x10, 0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, // IPv6 TE Router ID TLV
			},
			expectErr: false,
			expected: &RouterCapabilityTLV{
				TLVType:   242,
				TLVLength: 44,
				RouterID:  0x0a000100,
				Flags:     0,
				SubTLVs: []TLV{
					&SegmentRoutingCapabilityTLV{
						TLVType:   2,
						TLVLength: 9,
						Flags:     0xc0,
						Range:     [3]byte{0x00, 0x03, 0xe8},
						SubTLVs: []TLV{
							&SRCapSidLabelTLV{
								TLVType:   1,
								TLVLength: 3,
								Label:     [3]byte{0, 0x13, 0x88},
							},
						},
					},
					&SegmentRoutingAlgorithmsTLV{
						TLVType:   19,
						TLVLength: 2,
						Algorithms: []uint8{
							0,
							1,
						},
					},
					&NodeMaximumSIDDepthTLV{
						TLVType:   23,
						TLVLength: 4,
						MSDs: []MSD{
							{
								Type:  1,
								Value: 3,
							},
							{
								Type:  2,
								Value: 16,
							},
						},
					},
					&IPv6TERouterIDTLV{
						TLVType:   12,
						TLVLength: 16,
						RouterID: [16]byte{
							0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00,
							0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tt.input)
			result, err := readRouterCapabilityTLV(buf, 242, uint8(len(tt.input)))
			if tt.expectErr && err == nil {
				t.Errorf("expected error but got none for test case %s", tt.name)
				return
			}

			if !tt.expectErr && err != nil {
				t.Errorf("did not expect error but got: %q for test case %s", err, tt.name)
				return
			}

			assert.Equal(t, tt.expected, result, tt.name)
		})
	}
}

func TestRouterCapabilityTLVSerialize(t *testing.T) {
	tests := []struct {
		name     string
		tlv      *RouterCapabilityTLV
		expected []byte
	}{
		{
			name: "serialize without sub TLVs",
			tlv: &RouterCapabilityTLV{
				TLVType:   242,
				TLVLength: 7,
				RouterID:  0x01020304,
				Flags:     5,
				SubTLVs:   nil,
			},
			expected: []byte{242, 7, 1, 2, 3, 4, 5},
		},
		{
			name: "serialize with sub TLVs",
			tlv: &RouterCapabilityTLV{
				TLVType:   242,
				TLVLength: 10,
				RouterID:  0x01020304,
				Flags:     5,
				SubTLVs: []TLV{
					&UnknownTLV{TLVType: 2, TLVLength: 1, TLVValue: []byte{3}},
				},
			},
			expected: []byte{242, 10, 1, 2, 3, 4, 5, 2, 1, 3},
		},
		{
			name: "real world input from junos with SR enabled",
			tlv: &RouterCapabilityTLV{
				TLVType:   242,
				TLVLength: 44,
				RouterID:  0x0a000100,
				Flags:     0,
				SubTLVs: []TLV{
					&SegmentRoutingCapabilityTLV{
						TLVType:   2,
						TLVLength: 9,
						Flags:     0xc0,
						Range:     [3]byte{0x00, 0x03, 0xe8},
						SubTLVs: []TLV{
							&SRCapSidLabelTLV{
								TLVType:   1,
								TLVLength: 3,
								Label:     [3]byte{0, 0x13, 0x88},
							},
						},
					},
					&SegmentRoutingAlgorithmsTLV{
						TLVType:   19,
						TLVLength: 2,
						Algorithms: []uint8{
							0,
							1,
						},
					},
					&NodeMaximumSIDDepthTLV{
						TLVType:   23,
						TLVLength: 4,
						MSDs: []MSD{
							{
								Type:  1,
								Value: 3,
							},
							{
								Type:  2,
								Value: 16,
							},
						},
					},
					&IPv6TERouterIDTLV{
						TLVType:   12,
						TLVLength: 16,
						RouterID: [16]byte{
							0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00,
							0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02,
						},
					},
				},
			},
			expected: []byte{
				0xf2,                   // Router Capability TLV
				0x2c,                   // Length
				0x0a, 0x00, 0x01, 0x00, // Router ID
				0x00,       // Flags
				0x02, 0x09, // SR capability TLV
				0xc0,             // SR capability TLV flags
				0x00, 0x03, 0xe8, // SR capability TLV range
				0x01, 0x03, 0x00, 0x13, 0x88, // SID/Label TLV
				0x13, 0x02, 0x00, 0x01, // SR alrotighms TLV
				0x17, 0x04, 0x01, 0x03, 0x02, 0x10, // Node Maximum SID Depth TLV
				0x0c, 0x10, 0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, // IPv6 TE Router ID TLV
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			tt.tlv.Serialize(buf)
			result := buf.Bytes()

			assert.Equal(t, tt.expected, result, tt.name)
		})
	}
}
