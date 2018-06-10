package packet

type LSPTLV struct {
	TLVType    uint8
	TLVLength  uint8
	LSPEntries []LSPEntry
}

type LSPEntry struct {
	RemainingLifetime uint16
	LSPID             uint64
	LSPSequenceNumber uint32
	LSPChecksum       uint16
}
