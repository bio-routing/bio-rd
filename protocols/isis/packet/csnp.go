package packet

type CSNP struct {
	PDULength  uint16
	SourceID   [6]byte
	StartLSPID uint64
	EndLSPID   uint64
	LSPEntries []LSPEntry
}
