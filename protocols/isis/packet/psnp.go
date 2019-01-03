package packet

import (
	"bytes"
	"math"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/decode"
	umath "github.com/bio-routing/bio-rd/util/math"
	"github.com/pkg/errors"
	"github.com/taktv6/tflow2/convert"
)

const (
	// PSNPMinLen is the minimal length of PSNP PDU
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
		return nil, errors.Wrap(err, "Unable to decode fields")
	}

	nEntries := (psnp.PDULength - PSNPMinLen) / LSPEntryLen
	psnp.LSPEntries = make([]LSPEntry, nEntries)
	for i := uint16(0); i < nEntries; i++ {
		lspEntry, err := decodeLSPEntry(buf)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to get LSPEntries")
		}
		psnp.LSPEntries[i] = *lspEntry
	}

	return psnp, nil
}
