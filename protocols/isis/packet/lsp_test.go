package packet

import(
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestSerializeLSPDU(t *testing.T) {
	tests := []struct{
		name string
		lspdu *LSPDU
		expected []byte
	}{
		{
			name: "Test without TLVs",
			lspdu: &LSPDU{
				Length: 512,
				RemainingLifetime: 255,
				LSPID: LSPID{
					SystemID: types.SystemID{1, 2, 3, 4, 5, 6},
					PseudonodeID: 0,
				},
				SequenceNumber: 200,
				Checksum: 100,
				TypeBlock: 55,
				TLVs: make([]TLV, 0),
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
				Length: 512,
				RemainingLifetime: 255,
				LSPID: LSPID{
					SystemID: types.SystemID{1, 2, 3, 4, 5, 6},
					PseudonodeID: 0,
				},
				SequenceNumber: 200,
				Checksum: 100,
				TypeBlock: 55,
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
	tests := []struct{
		name string
		lspdu *LSPDU
		expected []byte
	}{
		{
			name: "Test without TLVs",
			lspdu: &LSPDU{
				Length: 512,
				RemainingLifetime: 255,
				LSPID: LSPID{
					SystemID: types.SystemID{1, 2, 3, 4, 5, 6},
					PseudonodeID: 0,
				},
				SequenceNumber: 200,
				TypeBlock: 55,
				TLVs: make([]TLV, 0),
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
				Length: 512,
				RemainingLifetime: 255,
				LSPID: LSPID{
					SystemID: types.SystemID{1, 2, 3, 4, 5, 6},
					PseudonodeID: 0,
				},
				SequenceNumber: 200,
				TypeBlock: 55,
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