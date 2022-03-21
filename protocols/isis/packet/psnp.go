package packet

import (
	"bytes"
	"fmt"
	"math"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/decode"
	umath "github.com/bio-routing/bio-rd/util/math"
	"github.com/bio-routing/tflow2/convert"
)

const (
	// PSNPMinLen is the minimal length of PSNP PDU
	PSNPMinLen = 17
)

// PSNP represents a Partial Sequence Number PDU
type PSNP struct {
	PDULength uint16
	SourceID  types.SourceID
	TLVs      []TLV
}

// NewPSNPs creates the necessary number of PSNP PDUs to carry all LSPEntries
func NewPSNPs(sourceID types.SourceID, lspEntries []*LSPEntry, maxPDULen int) []PSNP {
	left := len(lspEntries)
	lspsPerPSNP := (maxPDULen - PSNPMinLen - 2) / LSPEntryLen
	numPSNPs := int(math.Ceil(float64(left) / float64(lspsPerPSNP)))
	res := make([]PSNP, numPSNPs)

	for i := 0; i < numPSNPs; i++ {
		start := i * lspsPerPSNP
		end := umath.Min(lspsPerPSNP, left)

		slice := lspEntries[start : start+end]
		PSNP := newPSNP(sourceID, slice)
		if PSNP == nil {
			continue
		}

		res[i] = *PSNP
	}

	return res
}

func newPSNP(sourceID types.SourceID, lspEntries []*LSPEntry) *PSNP {
	if len(lspEntries) == 0 {
		return nil
	}

	psnp := PSNP{
		PDULength: PSNPMinLen + 2 + uint16(len(lspEntries))*LSPEntryLen,
		SourceID:  sourceID,
		TLVs: []TLV{
			NewLSPEntriesTLV(lspEntries),
		},
	}

	return &psnp
}

// GetLSPEntries returns LSP Entries from the LSP Entries TLV
func (p *PSNP) GetLSPEntries() []*LSPEntry {
	return getLSPEntries(p.TLVs)
}

// Serialize serializes PSNPs
func (p *PSNP) Serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint16Byte(p.PDULength))
	buf.Write(p.SourceID.Serialize())
	for _, tlv := range p.TLVs {
		tlv.Serialize(buf)
	}
}

// DecodePSNP decodes a Partion Sequence Number PDU
func DecodePSNP(buf *bytes.Buffer) (*PSNP, error) {
	psnp := &PSNP{}

	fields := []interface{}{
		&psnp.PDULength,
		&psnp.SourceID,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	tlvs, err := readTLVs(buf)
	if err != nil {
		return nil, fmt.Errorf("unable to read TLVs: %w", err)
	}

	psnp.TLVs = tlvs
	return psnp, nil
}
