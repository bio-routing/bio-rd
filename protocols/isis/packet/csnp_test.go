package packet

import (
	"bytes"
	"testing"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/stretchr/testify/assert"
)

func TestGetLSPEntries(t *testing.T) {
	tests := []struct {
		name     string
		csnp     *CSNP
		expected []*LSPEntry
	}{
		{
			name: "Test #1",
			csnp: &CSNP{
				TLVs: []TLV{
					&LSPEntriesTLV{
						TLVType: LSPEntriesTLVType,
						LSPEntries: []*LSPEntry{
							{
								SequenceNumber: 123,
							},
						},
					},
				},
			},
			expected: []*LSPEntry{
				{
					SequenceNumber: 123,
				},
			},
		},
		{
			name: "TLV not found",
			csnp: &CSNP{
				TLVs: []TLV{},
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		res := test.csnp.GetLSPEntries()
		assert.Equalf(t, test.expected, res, "Test %q", test.name)
	}
}

func TestNewCSNPs(t *testing.T) {
	tests := []struct {
		name         string
		sourceID     types.SourceID
		lspEntries   []*LSPEntry
		maxPDULength int
		expected     []CSNP
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
			expected: []CSNP{
				{
					PDULength: 49,
					SourceID: types.SourceID{
						SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
						CircuitID: 0,
					},
					StartLSPID: LSPID{},
					EndLSPID: LSPID{
						SystemID:     types.SystemID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						PseudonodeID: 0xffff,
					},
					TLVs: []TLV{
						&LSPEntriesTLV{
							TLVType:   LSPEntriesTLVType,
							TLVLength: 16,
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
					},
				},
				{
					SequenceNumber:    1000,
					RemainingLifetime: 2000,
					LSPChecksum:       111,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 100,
					},
				},
			},
			maxPDULength: 49,
			expected: []CSNP{
				{
					PDULength: 49,
					SourceID: types.SourceID{
						SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
						CircuitID: 0,
					},
					StartLSPID: LSPID{},
					EndLSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 100,
					},
					TLVs: []TLV{
						&LSPEntriesTLV{
							TLVType:   LSPEntriesTLVType,
							TLVLength: 16,
							LSPEntries: []*LSPEntry{
								{
									SequenceNumber:    1000,
									RemainingLifetime: 2000,
									LSPChecksum:       111,
									LSPID: LSPID{
										SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
										PseudonodeID: 100,
									},
								},
							},
						},
					},
				},
				{
					PDULength: 49,
					SourceID: types.SourceID{
						SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
						CircuitID: 0,
					},
					StartLSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 200,
					},
					EndLSPID: LSPID{
						SystemID:     types.SystemID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						PseudonodeID: 0xffff,
					},
					TLVs: []TLV{
						&LSPEntriesTLV{

							TLVType:   LSPEntriesTLVType,
							TLVLength: 16,
							LSPEntries: []*LSPEntry{
								{
									SequenceNumber:    1001,
									RemainingLifetime: 2001,
									LSPChecksum:       112,
									LSPID: LSPID{
										SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
										PseudonodeID: 200,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "2 packets with odd pdu length",
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
					},
				},
				{
					SequenceNumber:    1000,
					RemainingLifetime: 2000,
					LSPChecksum:       111,
					LSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 100,
					},
				},
			},
			maxPDULength: 55,
			expected: []CSNP{
				{
					PDULength: 49,
					SourceID: types.SourceID{
						SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
						CircuitID: 0,
					},
					StartLSPID: LSPID{},
					EndLSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 100,
					},
					TLVs: []TLV{
						&LSPEntriesTLV{
							TLVType:   LSPEntriesTLVType,
							TLVLength: 16,
							LSPEntries: []*LSPEntry{
								{
									SequenceNumber:    1000,
									RemainingLifetime: 2000,
									LSPChecksum:       111,
									LSPID: LSPID{
										SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
										PseudonodeID: 100,
									},
								},
							},
						},
					},
				},
				{
					PDULength: 49,
					SourceID: types.SourceID{
						SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
						CircuitID: 0,
					},
					StartLSPID: LSPID{
						SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
						PseudonodeID: 200,
					},
					EndLSPID: LSPID{
						SystemID:     types.SystemID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						PseudonodeID: 0xffff,
					},
					TLVs: []TLV{
						&LSPEntriesTLV{
							TLVType:   LSPEntriesTLVType,
							TLVLength: 16,
							LSPEntries: []*LSPEntry{
								{
									SequenceNumber:    1001,
									RemainingLifetime: 2001,
									LSPChecksum:       112,
									LSPID: LSPID{
										SystemID:     types.SystemID{10, 20, 30, 40, 50, 60},
										PseudonodeID: 200,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		csnps := NewCSNPs(test.sourceID, test.lspEntries, test.maxPDULength)
		assert.Equalf(t, test.expected, csnps, "Test: %q", test.name)
	}
}

func TestCSNPSerialize(t *testing.T) {
	tests := []struct {
		name     string
		csnp     CSNP
		expected []byte
	}{
		{
			name: "Test #1",
			csnp: CSNP{
				PDULength: 43,
				SourceID: types.SourceID{
					SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
					CircuitID: 0,
				},
				StartLSPID: LSPID{
					SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
					PseudonodeID: 256,
				},
				EndLSPID: LSPID{
					SystemID:     types.SystemID{11, 22, 33, 44, 55, 67},
					PseudonodeID: 255,
				},
				TLVs: []TLV{
					&LSPEntriesTLV{
						TLVType:   LSPEntriesTLVType,
						TLVLength: 16,
						LSPEntries: []*LSPEntry{
							{
								SequenceNumber:    123,
								RemainingLifetime: 255,
								LSPChecksum:       111,
								LSPID: LSPID{
									SystemID:     types.SystemID{10, 20, 30, 40, 50, 61},
									PseudonodeID: 11,
								},
							},
						},
					},
				},
			},
			expected: []byte{
				0, 43,
				10, 20, 30, 40, 50, 60, 0,
				11, 22, 33, 44, 55, 66, 1, 0,
				11, 22, 33, 44, 55, 67, 0, 255,
				9,
				16,
				0, 255,
				10, 20, 30, 40, 50, 61, 0, 11,
				0, 0, 0, 123,
				0, 111,
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(nil)
		test.csnp.Serialize(buf)
		assert.Equalf(t, test.expected, buf.Bytes(), "Test %q", test.name)
	}
}

func TestDecodeCSNP(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantFail bool
		expected *CSNP
	}{
		{
			name: "Incomplete CSNP",
			input: []byte{
				0, 25, // Length
				10, 20, 30, 40, 50, 60, 0, // Source ID
				11, 22, 33, 44, 55, 66, 0, 100,
				11, 22, 33, 77, 88, 0, 0,
			},
			wantFail: true,
		},
		{
			name: "Incomplete CSNP LSPEntry",
			input: []byte{
				0, 41, // Length
				10, 20, 30, 40, 50, 60, 0, // Source ID
				0, 0, 0, 20, // Sequence Number
				11, 22, 33, 44, 55, 66, 0, 100,
				11, 22, 33, 77, 88, 0, 0, 200,
				0, 0, 0, 20, // Sequence Number
			},
			wantFail: true,
		},
		{
			name: "CSNP with one LSPEntry",
			input: []byte{
				0, 43, // Length
				10, 20, 30, 40, 50, 60, 0, // Source ID
				11, 22, 33, 44, 55, 66, 0, 100, // StartLSPID
				11, 22, 33, 77, 88, 0, 0, 200, // EndLSPID
				9,    // TLV Type
				16,   // TLV Length
				1, 0, // Remaining Lifetime
				11, 22, 33, 44, 55, 66, // SystemID
				0, 20, // Pseudonode ID
				0, 0, 0, 20, // Sequence Number
				2, 0, // Checksum
			},
			wantFail: false,
			expected: &CSNP{
				PDULength: 43,
				SourceID: types.SourceID{
					SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
					CircuitID: 0,
				},
				StartLSPID: LSPID{
					SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
					PseudonodeID: 100,
				},
				EndLSPID: LSPID{
					SystemID:     types.SystemID{11, 22, 33, 77, 88, 0},
					PseudonodeID: 200,
				},
				TLVs: []TLV{
					&LSPEntriesTLV{
						TLVType:   LSPEntriesTLVType,
						TLVLength: 16,
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
						},
					},
				},
			},
		},
		{
			name: "PSNP with two LSPEntries",
			input: []byte{
				0, 62, // Length
				10, 20, 30, 40, 50, 60, 0, // Source ID
				11, 22, 33, 44, 55, 66, 0, 100, // StartLSPID
				11, 22, 33, 77, 88, 0, 0, 200, // EndLSPID
				9,    // TLV Type
				32,   // TLV Length
				1, 0, // Remaining Lifetime
				11, 22, 33, 44, 55, 66, // SystemID
				0, 20, // Pseudonode ID
				0, 0, 0, 20, // Sequence Number
				2, 0, // Checksum
				2, 0, // Remaining Lifetime
				11, 22, 33, 44, 55, 67, // SystemID
				0, 21, // Pseudonode ID
				0, 0, 0, 21, // Sequence Number
				2, 0, // Checksum
			},
			wantFail: false,
			expected: &CSNP{
				PDULength: 62,
				SourceID: types.SourceID{
					SystemID:  types.SystemID{10, 20, 30, 40, 50, 60},
					CircuitID: 0,
				},
				StartLSPID: LSPID{
					SystemID:     types.SystemID{11, 22, 33, 44, 55, 66},
					PseudonodeID: 100,
				},
				EndLSPID: LSPID{
					SystemID:     types.SystemID{11, 22, 33, 77, 88, 0},
					PseudonodeID: 200,
				},
				TLVs: []TLV{
					&LSPEntriesTLV{
						TLVType:   LSPEntriesTLVType,
						TLVLength: 32,
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
			},
		},
	}

	for _, test := range tests {
		buf := bytes.NewBuffer(test.input)
		csnp, err := DecodeCSNP(buf)
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

		assert.Equalf(t, test.expected, csnp, "Test %q", test.name)
	}
}
