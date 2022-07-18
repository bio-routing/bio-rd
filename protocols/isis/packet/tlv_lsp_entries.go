package packet

import (
	"bytes"
	"fmt"
)

const (
	// LSPEntriesTLVType is the type value of an LSP Entries TLV
	LSPEntriesTLVType = uint8(9)
)

// LSPEntriesTLV is an LSP Entries TLV carried in PSNP/CSNP
type LSPEntriesTLV struct {
	TLVType    uint8
	TLVLength  uint8
	LSPEntries []*LSPEntry
}

func (l *LSPEntriesTLV) Copy() TLV {
	res := &LSPEntriesTLV{
		TLVType:    l.TLVType,
		TLVLength:  l.TLVLength,
		LSPEntries: make([]*LSPEntry, 0, len(l.LSPEntries)),
	}

	for _, e := range l.LSPEntries {
		c := *e
		res.LSPEntries = append(res.LSPEntries, &c)
	}

	return res
}

// Type returns the type of the TLV
func (l *LSPEntriesTLV) Type() uint8 {
	return l.TLVType
}

// Length returns the length of the TLV
func (l *LSPEntriesTLV) Length() uint8 {
	return l.TLVLength
}

// Value returns self
func (l *LSPEntriesTLV) Value() interface{} {
	return l
}

// NewLSPEntriesTLV creates a nbew LSP Entries TLV
func NewLSPEntriesTLV(LSPEntries []*LSPEntry) *LSPEntriesTLV {
	return &LSPEntriesTLV{
		TLVType:    LSPEntriesTLVType,
		TLVLength:  uint8(len(LSPEntries)) * LSPEntryLen,
		LSPEntries: LSPEntries,
	}
}

// Serialize serializes an LSP Entries TLV
func (l *LSPEntriesTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(l.TLVType)
	buf.WriteByte(l.TLVLength)
	for i := range l.LSPEntries {
		l.LSPEntries[i].Serialize(buf)
	}
}

func readLSPEntriesTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*LSPEntriesTLV, error) {
	pdu := &LSPEntriesTLV{
		TLVType:    tlvType,
		TLVLength:  tlvLength,
		LSPEntries: make([]*LSPEntry, 0, tlvLength/LSPEntryLen),
	}

	toRead := tlvLength
	for toRead > 0 {
		e, err := decodeLSPEntry(buf)
		if err != nil {
			return nil, fmt.Errorf("unable to decode LSP Entry: %w", err)
		}

		pdu.LSPEntries = append(pdu.LSPEntries, e)
		toRead -= LSPEntryLen
	}

	return pdu, nil
}
