package packet

import (
	"bytes"
	"fmt"
	"math"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/taktv6/tflow2/convert"
)

const (
	PSNPMinLen = 8
)

// PSNP represents a Partial Sequence Number PDU
type PSNP struct {
	PDULength  uint16
	SourceID   types.SystemID
	LSPEntries []LSPEntry
}

// NewPSNPs creates the necessary number of PSNP PDUs to carry all LSPEntries
func NewPSNPs(sourceID types.SystemID, lspEntries []LSPEntry, maxPDULen int) []PSNP {
	left := len(lspEntries)
	lspsPerPSNP := (maxPDULen - PSNPMinLen) / LSPEntryLen
	numPSNPs := int(math.Ceil(float64(left) / float64(lspsPerPSNP)))
	res := make([]PSNP, numPSNPs)

	for i := 0; i < numPSNPs; i++ {
		start := i * lspsPerPSNP
		end := min(lspsPerPSNP, left)

		slice := lspEntries[start : start+end]
		PSNP := newPSNP(sourceID, slice)
		if PSNP == nil {
			continue
		}

		res[i] = *PSNP
	}

	return res
}

func newPSNP(sourceID types.SystemID, lspEntries []LSPEntry) *PSNP {
	if len(lspEntries) == 0 {
		return nil
	}

	psnp := PSNP{
		PDULength:  PSNPMinLen + uint16(len(lspEntries))*LSPEntryLen,
		SourceID:   sourceID,
		LSPEntries: lspEntries,
	}

	return &psnp
}

// Serialize serializes PSNPs
func (c *PSNP) Serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint16Byte(c.PDULength))
	buf.Write(c.SourceID[:])

	for _, lspEntry := range c.LSPEntries {
		lspEntry.Serialize(buf)
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
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	nEntries := (psnp.PDULength - PSNPMinLen) / LSPEntryLen
	psnp.LSPEntries = make([]LSPEntry, nEntries)
	for i := uint16(0); i < nEntries; i++ {
		lspEntry, err := decodeLSPEntry(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to get LSPEntries: %v", err)
		}
		psnp.LSPEntries[i] = *lspEntry
	}

	return psnp, nil
}
