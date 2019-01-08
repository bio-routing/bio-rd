package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestNewPSNPs(t *testing.T) {
	tests := []struct {
		name         string
		sourceID     types.SourceID
		lspEntries   []*LSPEntry
		maxPDULength int
		expected     []PSNP
	}{
		{
			name: "All in one packet",
			sourceID: types.SourceID{
				SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
				CircuitID: 0,
			},
			lspEntries: []*LSPEntry{
				{
					SequenceNumber:    1000,
					RemainingLifetime: 2000,
					LSPChecksum:       111,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 123,
					},
				},
			},
			maxPDULength: 1492,
			expected: []PSNP{
				{
					PDULength: 35,
					SourceID: types.SourceID{
						SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
						CircuitID: 0,
					},
					LSPEntries: []*LSPEntry{
						{
							SequenceNumber:    1000,
							RemainingLifetime: 2000,
							LSPChecksum:       111,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 123,
							},
						},
					},
				},
			},
		},
		{
			name: "2 packets",
			sourceID: types.SourceID{
				SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
				CircuitID: 0,
			},
			lspEntries: []*LSPEntry{
				{
					SequenceNumber:    1001,
					RemainingLifetime: 2001,
					LSPChecksum:       112,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 200,
						LSPNumber:    100,
					},
				},
				{
					SequenceNumber:    1000,
					RemainingLifetime: 2000,
					LSPChecksum:       111,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 100,
						LSPNumber:    200,
					},
				},
			},
			maxPDULength: 35,
			expected: []PSNP{
				{
					PDULength: 35,
					SourceID: types.SourceID{
						SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
						CircuitID: 0,
					},
					LSPEntries: []*LSPEntry{
						{
							SequenceNumber:    1001,
							RemainingLifetime: 2001,
							LSPChecksum:       112,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 200,
								LSPNumber:    100,
							},
						},
					},
				},
				{
					PDULength: 35,
					SourceID: types.SourceID{
						SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
						CircuitID: 0,
					},
					LSPEntries: []*LSPEntry{
						{
							SequenceNumber:    1000,
							RemainingLifetime: 2000,
							LSPChecksum:       111,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 100,
								LSPNumber:    200,
							},
						},
					},
				},
			},
		},
		{
			name: "2 packets with odd length",
			sourceID: types.SourceID{
				SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
				CircuitID: 0,
			},
			lspEntries: []*LSPEntry{
				{
					SequenceNumber:    1001,
					RemainingLifetime: 2001,
					LSPChecksum:       112,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 200,
						LSPNumber:    10,
					},
				},
				{
					SequenceNumber:    1000,
					RemainingLifetime: 2000,
					LSPChecksum:       111,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 100,
						LSPNumber:    20,
					},
				},
			},
			maxPDULength: 40,
			expected: []PSNP{
				{
					PDULength: 35,
					SourceID: types.SourceID{
						SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
						CircuitID: 0,
					},
					LSPEntries: []*LSPEntry{
						{
							SequenceNumber:    1001,
							RemainingLifetime: 2001,
							LSPChecksum:       112,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 200,
								LSPNumber:    10,
							},
						},
					},
				},
				{
					PDULength: 35,
					SourceID: types.SourceID{
						SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
						CircuitID: 0,
					},
					LSPEntries: []*LSPEntry{
						{
							SequenceNumber:    1000,
							RemainingLifetime: 2000,
							LSPChecksum:       111,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 100,
								LSPNumber:    20,
							},
						},
					},
				},
			},
		},
		{
			name: "2 LSPEntries, 1 packet",
			sourceID: types.SourceID{
				SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
				CircuitID: 0,
			},
			lspEntries: []*LSPEntry{
				{
					SequenceNumber:    1001,
					RemainingLifetime: 2001,
					LSPChecksum:       112,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 200,
						LSPNumber:    10,
					},
				},
				{
					SequenceNumber:    1000,
					RemainingLifetime: 2000,
					LSPChecksum:       111,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 100,
						LSPNumber:    20,
					},
				},
			},
			maxPDULength: 51,
			expected: []PSNP{
				{
					PDULength: 51,
					SourceID: types.SourceID{
						SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
						CircuitID: 0,
					},
					LSPEntries: []*LSPEntry{
						{
							SequenceNumber:    1001,
							RemainingLifetime: 2001,
							LSPChecksum:       112,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 200,
								LSPNumber:    10,
							},
						},
						{
							SequenceNumber:    1000,
							RemainingLifetime: 2000,
							LSPChecksum:       111,
							LSPID: LSPID{
								SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
								PseudonodeID: 100,
								LSPNumber:    20,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		psnps := NewPSNPs(test.sourceID, test.lspEntries, test.maxPDULength)
		assert.Equalf(t, test.expected, psnps, "Test %q", test.name)
	}
}

func TestPSNPSerialize(t *testing.T) {
	tests := []struct {
		name     string
		psnp     PSNP
		expected []byte
	}{
		{
			name: "Test #2",
			psnp: PSNP{
				PDULength: 0x23,
				SourceID: types.SourceID{
					SystemID:  types.SystemID{0, 0, 0, 0, 0, 3},
					CircuitID: 0,
				},
				LSPEntries: []*LSPEntry{
					{
						SequenceNumber:    0x15,
						RemainingLifetime: 1196,
						LSPChecksum:       0xe4ef,
						LSPID: LSPID{
							SystemID:     types.SystemID{0, 0, 0, 0, 0, 2},
							PseudonodeID: 0,
						},
					},
				},
			},
			expected: []byte{
				0x00, 0x23, // Length
				0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x00, // SystemID
				0x09,       // TLV Type
				0x10,       // TLV Length
				0x04, 0xac, // Remaining Lifetime
				0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, //LSPID
				0x00, 0x00, 0x00, 0x15, // Sequence Number
				0xe4, 0xef, // // Checksum
			},
		},
		{
			name: "Test #1",
			psnp: PSNP{
				PDULength: 100,
				SourceID: types.SourceID{
					SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
					CircuitID: 0,
				},
				LSPEntries: []*LSPEntry{
					{
						RemainingLifetime: 255,
						LSPID: LSPID{
							SystemID:     types.SystemID{10, 20, 30, 40, 50, 61},
							PseudonodeID: 11,
							LSPNumber:    10,
						},
						SequenceNumber: 123,
						LSPChecksum:    111,
					},
				},
			},
			expected: []byte{
				0, 100, // Length
				10, 20, 30, 40, 50, 60, 0, // SystemID
				9,      // TLV Type
				16,     // TLV Length
				0, 255, // Remaining Lifetime
				10, 20, 30, 40, 50, 61, 11, 10, // LSPID
				0, 0, 0, 123, // Sequence Number
				0, 111, // Checksum
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.psnp.Serialize(buf)
		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}

func TestDecodePSNP(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *PSNP
	}{
		{
			name: "Incomplete PSNP",
			input: []byte{
				0, 33, // Length
				10, 20, 30, 40, 50, 60, 0, // Source ID
			},
			wantFail: true,
		},
		{
			name: "Incomplete PSNP LSPEntry",
			input: []byte{
				0, 33, // Length
				10, 20, 30, 40, 50, 60, 0, // Source ID
				0, 0, 0, 20, // Sequence Number
			},
			wantFail: true,
		},
		{
			name: "PSNP with one LSPEntry",
			input: []byte{
				0, 33, // Length
				10, 20, 30, 40, 50, 60, 0, // Source ID

				1, 0, // Remaining Lifetime
				11, 22, 33, 44, 55, 66, // SystemID
				0,           // Pseudonode ID
				20,          // LSPNumber
				0, 0, 0, 20, // Sequence Number
				2, 0, // Checksum

			},
			wantFail: false,
			expected: &PSNP{
				PDULength: 33,
				SourceID: types.SourceID{
					SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
					CircuitID: 0,
				},
				LSPEntries: []*LSPEntry{
					{
						SequenceNumber:    20,
						RemainingLifetime: 256,
						LSPChecksum:       512,
						LSPID: LSPID{
							SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
							PseudonodeID: 0,
							LSPNumber:    20,
						},
					},
				},
			},
		},
		{
			name: "PSNP with two LSPEntries",
			input: []byte{
				0, 49, // Length
				10, 20, 30, 40, 50, 60, 0, // Source ID

				1, 0, // Remaining Lifetime
				11, 22, 33, 44, 55, 66, // SystemID
				20,          // Pseudonode ID
				0,           // LSP Number
				0, 0, 0, 20, // Sequence Number
				2, 0, // Checksum

				2, 0, // Remaining Lifetime
				11, 22, 33, 44, 55, 67, // SystemID
				21,          // Pseudonode ID
				00,          // LSP Number
				0, 0, 0, 21, // Sequence Number
				2, 0, // Checksum
			},
			wantFail: false,
			expected: &PSNP{
				PDULength: 49,
				SourceID: types.SourceID{
					SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
					CircuitID: 0,
				},
				LSPEntries: []*LSPEntry{
					{
						SequenceNumber:    20,
						RemainingLifetime: 256,
						LSPChecksum:       512,
						LSPID: LSPID{
							SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
							PseudonodeID: 20,
						},
					},
					{
						SequenceNumber:    21,
						RemainingLifetime: 512,
						LSPChecksum:       512,
						LSPID: LSPID{
							SystemID:     types.SystemID{11, 22, 33, 44, 55, 67},
							PseudonodeID: 21,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		psnp, err := DecodePSNP(buf)
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

		assert.Equalf(t, test.expected, psnp, "Test %q", test.name)
	}
}
