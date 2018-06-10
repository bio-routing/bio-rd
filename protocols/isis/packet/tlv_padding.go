package packet

const PaddingType = 8

type PaddingTLV struct {
	TLVType     uint8
	TLVLength   uint8
	PaddingData []byte
}
