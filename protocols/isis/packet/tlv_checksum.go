package packet

const ChecksumType = 12

type ChecksumTLV struct {
	TLVType   uint8
	TLVLength uint8
	Checksum  uint16
}
