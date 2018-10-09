package packet

import (
	"bytes"
	"fmt"
	"math"
	"sort"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/decode"
	umath "github.com/bio-routing/bio-rd/util/math"
	"github.com/taktv6/tflow2/convert"
)

const (
	// CSNPMinLen is the minimal length of a CSNP
	CSNPMinLen = 24
)

// CSNP represents a Complete Sequence Number PDU
type CSNP struct {
	PDULength  uint16
	SourceID   [6]byte
	StartLSPID LSPID
	EndLSPID   LSPID
	LSPEntries []LSPEntry
}

func compareLSPIDs(lspIDA, lspIDB LSPID) bool {
	for i := 0; i < len(lspIDA.SystemID); i++ {
		if lspIDA.SystemID[i] < lspIDB.SystemID[i] {
			return true
		}
		if lspIDA.SystemID[i] > lspIDB.SystemID[i] {
			return false
		}
	}

	if lspIDA.PseudonodeID < lspIDB.PseudonodeID {
		return true
	}

	return false
}

// NewCSNPs creates the necessary number of CSNP PDUs to carry all LSPEntries
func NewCSNPs(sourceID types.SystemID, lspEntries []LSPEntry, maxPDULen int) []CSNP {
	left := len(lspEntries)
	lspsPerCSNP := (maxPDULen - CSNPMinLen) / LSPEntryLen
	numCSNPs := int(math.Ceil(float64(left) / float64(lspsPerCSNP)))
	res := make([]CSNP, numCSNPs)

	sort.Slice(lspEntries, func(a, b int) bool {
		for i := 0; i < len(lspEntries[a].LSPID.SystemID); i++ {
			if lspEntries[a].LSPID.SystemID[i] < lspEntries[b].LSPID.SystemID[i] {
				return true
			}
			if lspEntries[a].LSPID.SystemID[i] > lspEntries[b].LSPID.SystemID[i] {
				return false
			}
		}

		if lspEntries[a].LSPID.PseudonodeID < lspEntries[b].LSPID.PseudonodeID {
			return true
		}

		return false
	})

	for i := 0; i < numCSNPs; i++ {
		start := i * lspsPerCSNP
		end := umath.Min(lspsPerCSNP, left)

		slice := lspEntries[start : start+end]
		csnp := newCSNP(sourceID, slice)
		if csnp == nil {
			continue
		}

		res[i] = *csnp
	}

	res[0].StartLSPID = LSPID{}
	res[len(res)-1].EndLSPID = LSPID{
		SystemID:     types.SystemID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		PseudonodeID: 0xffff,
	}

	return res
}

func newCSNP(sourceID types.SystemID, lspEntries []LSPEntry) *CSNP {
	if len(lspEntries) == 0 {
		return nil
	}

	csnp := CSNP{
		PDULength:  uint16(CSNPMinLen + len(lspEntries)*LSPEntryLen),
		SourceID:   sourceID,
		StartLSPID: lspEntries[0].LSPID,
		EndLSPID:   lspEntries[len(lspEntries)-1].LSPID,
		LSPEntries: lspEntries,
	}

	return &csnp
}

// Serialize serializes CSNPs
func (c *CSNP) Serialize(buf *bytes.Buffer) {
	c.PDULength = uint16(CSNPMinLen + len(c.LSPEntries)*LSPEntryLen)
	buf.Write(convert.Uint16Byte(c.PDULength))
	buf.Write(c.SourceID[:])
	c.StartLSPID.Serialize(buf)
	c.EndLSPID.Serialize(buf)

	for _, lspEntry := range c.LSPEntries {
		lspEntry.Serialize(buf)
	}
}

// DecodeCSNP decodes Complete Sequence Number PDUs
func DecodeCSNP(buf *bytes.Buffer) (*CSNP, error) {
	csnp := &CSNP{}

	fields := []interface{}{
		&csnp.PDULength,
		&csnp.SourceID,
		&csnp.StartLSPID.SystemID,
		&csnp.StartLSPID.PseudonodeID,
		&csnp.EndLSPID.SystemID,
		&csnp.EndLSPID.PseudonodeID,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	nEntries := (csnp.PDULength - CSNPMinLen) / LSPEntryLen
	csnp.LSPEntries = make([]LSPEntry, nEntries)
	for i := uint16(0); i < nEntries; i++ {
		lspEntry, err := decodeLSPEntry(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to get LSPEntries: %v", err)
		}
		csnp.LSPEntries[i] = *lspEntry
	}

	return csnp, nil
}
