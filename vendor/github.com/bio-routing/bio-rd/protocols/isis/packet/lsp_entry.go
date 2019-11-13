package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"
)

const (
	// LSPEntryLen is the lenth of an LSPEntry
	LSPEntryLen = 16
)

// LSPEntry represents an LSP entry in a CSNP PDU
type LSPEntry struct {
	RemainingLifetime uint16
	LSPID             LSPID
	SequenceNumber    uint32
	LSPChecksum       uint16
}

// Serialize serializes an LSPEntry
func (l *LSPEntry) Serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint16Byte(l.RemainingLifetime))
	l.LSPID.Serialize(buf)
	buf.Write(convert.Uint32Byte(l.SequenceNumber))
	buf.Write(convert.Uint16Byte(l.LSPChecksum))
}

func decodeLSPEntry(buf *bytes.Buffer) (*LSPEntry, error) {
	lspEntry := &LSPEntry{}

	fields := []interface{}{
		&lspEntry.RemainingLifetime,
		&lspEntry.LSPID.SystemID,
		&lspEntry.LSPID.PseudonodeID,
		&lspEntry.LSPID.LSPNumber,
		&lspEntry.SequenceNumber,
		&lspEntry.LSPChecksum,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	return lspEntry, nil
}
