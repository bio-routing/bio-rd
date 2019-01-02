package packet

import (
	"bytes"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/pkg/errors"
	"github.com/bio-routing/tflow2/convert"
)

const (
	// LSPEntryLen is the lenth of an LSPEntry
	LSPEntryLen = 16
)

// LSPEntry represents an LSP entry in a CSNP PDU
type LSPEntry struct {
	SequenceNumber    uint32
	RemainingLifetime uint16
	LSPChecksum       uint16
	LSPID             LSPID
}

// Serialize serializes an LSPEntry
func (l *LSPEntry) Serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint32Byte(l.SequenceNumber))
	buf.Write(convert.Uint16Byte(l.RemainingLifetime))
	buf.Write(convert.Uint16Byte(l.LSPChecksum))
	l.LSPID.Serialize(buf)
}

func decodeLSPEntry(buf *bytes.Buffer) (*LSPEntry, error) {
	lspEntry := &LSPEntry{}

	fields := []interface{}{
		&lspEntry.SequenceNumber,
		&lspEntry.RemainingLifetime,
		&lspEntry.LSPChecksum,
		&lspEntry.LSPID.SystemID,
		&lspEntry.LSPID.PseudonodeID,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to decode fields")
	}

	return lspEntry, nil
}
