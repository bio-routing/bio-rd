package packet

const ExtendedIPReachabilityTLVType = 135

type ExtendedIPReachabilityTLV struct {
	TLVType        uint8
	TLVLength      uint8
	Metric         uint32
	UDSubBitPfxLen uint8
	Address        uint32
	SubTLVType     uint8
	SubTLVLength   uint8
	SubTLVs        []interface{}
}
