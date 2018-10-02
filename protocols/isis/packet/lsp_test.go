package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

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

func TestSetChecksum(t *testing.T) {
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
				TypeBlock:      55,
				TLVs:           make([]TLV, 0),
			},
			expected: []byte{
				2, 0,
				0, 255,
				1, 2, 3, 4, 5, 6, 0, 0,
				0, 0, 0, 200,
				0x76, 0x17,
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
				0xf8, 0x21,
				55,
				8, 2,
				0, 0,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.lspdu.SetChecksum()
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
