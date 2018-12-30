package packet

import "bytes"

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
