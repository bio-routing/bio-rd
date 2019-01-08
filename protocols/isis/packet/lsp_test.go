package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestLSPIDCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        LSPID
		b        LSPID
		expected int
	}{
		{
			name:     "Test #1",
			a:        LSPID{},
			b:        LSPID{},
			expected: 0,
		},
		{
			name: "Test #2",
			a: LSPID{
				SystemID:     types.SystemID{1, 2, 3, 4, 5, 6},
				PseudonodeID: 100,
			},
			b: LSPID{
				SystemID:     types.SystemID{1, 2, 3, 4, 5, 7},
				PseudonodeID: 100,
			},
			expected: -1,
		},
		{
			name: "Test #3",
			a: LSPID{
				SystemID:     types.SystemID{1, 2, 3, 4, 5, 8},
				PseudonodeID: 100,
			},
			b: LSPID{
				SystemID:     types.SystemID{1, 2, 3, 4, 5, 7},
				PseudonodeID: 100,
			},
			expected: 1,
		},
		{
			name: "Test #4",
			a: LSPID{
				SystemID:     types.SystemID{1, 2, 3, 4, 5, 7},
				PseudonodeID: 101,
			},
			b: LSPID{
				SystemID:     types.SystemID{1, 2, 3, 4, 5, 7},
				PseudonodeID: 100,
			},
			expected: 1,
		},
		{
			name: "Test #4",
			a: LSPID{
				SystemID:     types.SystemID{1, 2, 3, 4, 5, 7},
				PseudonodeID: 101,
			},
			b: LSPID{
				SystemID:     types.SystemID{1, 2, 3, 4, 5, 7},
				PseudonodeID: 102,
			},
			expected: -1,
		},
	}

	for _, test := range tests {
		res := test.a.Compare(test.b)
		assert.Equalf(t, test.expected, res, "Test %q", test.name)
	}
}

func TestSerializeLSPDU(t *testing.T) {
	tests := []struct {
		name     string
		lspdu    *LSPDU
		expected []byte
	}{
		{
			name: "Test without TLVs",
			lspdu: &LSPDU{
				Length:            512,
				RemainingLifetime: 255,
				LSPID: LSPID{
					SystemID:     types.SystemID{1, 2, 3, 4, 5, 6},
					PseudonodeID: 0,
				},
				SequenceNumber: 200,
				Checksum:       100,
				TypeBlock:      55,
				TLVs:           make([]TLV, 0),
			},
			expected: []byte{
				2, 0,
				0, 255,
				1, 2, 3, 4, 5, 6, 0, 0,
				0, 0, 0, 200,
				0, 100,
				55,
			},
		},
		{
			name: "Test with TLV",
			lspdu: &LSPDU{
				Length:            512,
				RemainingLifetime: 255,
				LSPID: LSPID{
					SystemID:     types.SystemID{1, 2, 3, 4, 5, 6},
					PseudonodeID: 0,
				},
				SequenceNumber: 200,
				Checksum:       100,
				TypeBlock:      55,
				TLVs: []TLV{
					NewPaddingTLV(2),
				},
			},
			expected: []byte{
				2, 0,
				0, 255,
				1, 2, 3, 4, 5, 6, 0, 0,
				0, 0, 0, 200,
				0, 100,
				55,
				8, 2,
				0, 0,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.lspdu.Serialize(buf)
		assert.Equalf(t, test.expected, buf.Bytes(), "Unexpected result in test %q", test.name)
	}
}

func TestDecodeLSPDU(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *LSPDU
	}{
		{
			name: "Incomplete LSPDU",
			input: []byte{
				0, 17, // Length
				0, 200, // Lifetime
				10, 20, 30, 40, 50, 60, 0, 10, // LSPID
				0, 0, 1, 0, // Sequence Number
			},
			wantFail: true,
		},
		{
			name: "Incomplete TLV",
			input: []byte{
				0, 25, // Length
				0, 200, // Lifetime
				10, 20, 30, 40, 50, 60, 0, 10, // LSPID
				0, 0, 1, 0, // Sequence Number
				0, 0, // Checksum
				137, 5, 1, 2, 3, 4, // Incomplete Hostname TLV
			},
			wantFail: true,
		},
		{
			name: "LSP with two TLVs",
			input: []byte{
				0, 29, // Length
				0, 200, // Lifetime
				10, 20, 30, 40, 50, 60, 0, 10, // LSPID
				0, 0, 1, 0, // Sequence Number
				0, 0, // Checksum
				3,                     // Typeblock
				137, 5, 1, 2, 3, 4, 5, // Hostname TLV
				12, 2, 0, 2, // Checksum TLV
			},
			wantFail: false,
			expected: &LSPDU{
				Length:            29,
				RemainingLifetime: 200,
				LSPID: LSPID{
					SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
					PseudonodeID: 10,
				},
				SequenceNumber: 256,
				Checksum:       0,
				TypeBlock:      3,
				TLVs: []TLV{
					&DynamicHostNameTLV{
						TLVType:   137,
						TLVLength: 5,
						Hostname:  []byte{1, 2, 3, 4, 5},
					},
					&ChecksumTLV{
						TLVType:   12,
						TLVLength: 2,
						Checksum:  2,
					},
				},
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		lspdu, err := DecodeLSPDU(buf)
		if err != nil {
			if test.wantFail {
				continue
			}
			t.Errorf("Unexpected failure for test %q: %v", test.name, err)
			continue
		}

		if test.wantFail {
			t.Errorf("Unexpected success for test %q", test.name)
			continue
		}

		assert.Equalf(t, test.expected, lspdu, "Test %q", test.name)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		input    *LSPID
		expected string
	}{
		{
			name: "Test #1",
			input: &LSPID{
				SystemID:     types.SystemID{1, 2, 3, 4, 5, 6},
				PseudonodeID: 5,
			},
			expected: "010203.040506-05",
		},
	}

	for _, test := range tests {
		res := test.input.String()
		assert.Equal(t, test.expected, res, test.name)
	}
}

func TestSetChecksum(t *testing.T) {
	tests := []struct {
		name     string
		lspdu    *LSPDU
		expected uint16
	}{
		{
			name: "Test #1",
			lspdu: &LSPDU{
				Length:            29,
				RemainingLifetime: 3591,
				LSPID: LSPID{
					SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
					PseudonodeID: 0,
					LSPNumber:    0,
				},
				SequenceNumber: 1,
				TypeBlock:      3,
				TLVs: []TLV{
					AreaAddressesTLV{
						TLVType:   AreaAddressesTLVType,
						TLVLength: 6,
						AreaIDs: []types.AreaID{
							{
								0x49, 0, 1, 0, 16,
							},
						},
					},
					ProtocolsSupportedTLV{
						TLVType:   ProtocolsSupportedTLVType,
						TLVLength: 2,
						NetworkLayerProtocolIDs: []uint8{
							0xcc, 0x8e,
						},
					},
				},
			},
			expected: 0x0b1e,
		},
	}

	for _, test := range tests {
		test.lspdu.SetChecksum()
		assert.Equal(t, test.expected, test.lspdu.Checksum, test.name)
	}
}
