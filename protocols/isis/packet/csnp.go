package packet

import(
	"bytes"

	"github.com/taktv6/tflow2/convert"
)

const(
	CSNPMinLen = 24
)

type CSNP struct {
	PDULength  uint16
	SourceID   [6]byte
	StartLSPID uint64
	EndLSPID   uint64
	LSPEntries []LSPEntry
}

func (c *CSNP) Serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint16Byte(c.PDULength))
	buf.Write(c.SourceID[:])
	buf.Write(convert.Uint64Byte(c.StartLSPID))
	buf.Write(convert.Uint64Byte(c.EndLSPID))

	for _, lspEntry := range c.LSPEntries {
		lspEntry.Serialize(buf)
	}
}

type LSPEntry struct {
	SequenceNumber uint32
	RemainingLifetime uint16
	LSPChecksum uint16
	LSPID LSPID
}

func (l *LSPEntry) Serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint32Byte(l.SequenceNumber))
	buf.Write(convert.Uint16Byte(l.RemainingLifetime))
	buf.Write(convert.Uint16Byte(l.LSPChecksum))
	l.LSPID.Serialize(buf)
}