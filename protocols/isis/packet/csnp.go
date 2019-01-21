package packet

import (
	"bytes"
	"math"
	"sort"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/decode"
	umath "github.com/bio-routing/bio-rd/util/math"
	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"
)

const (
	// CSNPMinLen is the minimal length of a CSNP
	CSNPMinLen = PSNPMinLen + 16
)

// CSNP represents a Complete Sequence Number PDU
type CSNP struct {
	PDULength  uint16
	SourceID   types.SourceID
	StartLSPID LSPID
	EndLSPID   LSPID
	TLVs       []TLV
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
func NewCSNPs(sourceID types.SourceID, lspEntries []*LSPEntry, maxPDULen int) []CSNP {
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

		entries := lspEntries[start : start+end]
		tlvs := []TLV{
			NewLSPEntriesTLV(entries),
		}
		csnp := newCSNP(sourceID, entries[0].LSPID, entries[len(entries)-1].LSPID, tlvs)
		if csnp == nil {
			continue
		}

		res[i] = *csnp
	}

	res[0].StartLSPID = LSPID{}
	res[len(res)-1].EndLSPID = LSPID{
		SystemID:     types.SystemID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		PseudonodeID: 0xff,
		LSPNumber:    0xff,
	}

	return res
}

func newCSNP(sourceID types.SourceID, startLSPID LSPID, endLSPID LSPID, tlvs []TLV) *CSNP {
	tlvsLen := uint16(0)
	for i := range tlvs {
		tlvsLen += uint16(tlvs[i].Length())
	}

	csnp := CSNP{
		PDULength:  uint16(CSNPMinLen + tlvsLen),
		SourceID:   sourceID,
		StartLSPID: startLSPID,
		EndLSPID:   endLSPID,
		TLVs:       tlvs,
	}

	return &csnp
}

// GetLSPEntries returns LSP Entries from the LSP Entries TLV
func (c *CSNP) GetLSPEntries() []*LSPEntry {
	for _, tlv := range c.TLVs {
		if tlv.Type() != LSPEntriesTLVType {
			continue
		}

		return tlv.Value().(*LSPEntriesTLV).LSPEntries
	}

	return nil
}

// Serialize serializes CSNPs
func (c *CSNP) Serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint16Byte(c.PDULength))
	buf.Write(c.SourceID.Serialize())
	c.StartLSPID.Serialize(buf)
	c.EndLSPID.Serialize(buf)

	for i := range c.TLVs {
		c.TLVs[i].Serialize(buf)
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
		&csnp.StartLSPID.LSPNumber,
		&csnp.EndLSPID.SystemID,
		&csnp.EndLSPID.PseudonodeID,
		&csnp.EndLSPID.LSPNumber,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to decode fields")
	}

	tlvs, err := readTLVs(buf)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read TLVs")
	}

	csnp.TLVs = tlvs
	return csnp, nil
}
