package packet

const ExtendedISReachabilityType = 22

type ExtendedISReachabilityTLV struct {
	TLVType      uint8
	TLVLength    uint8
	SystemID     [7]byte
	WideMetrics  [3]byte
	SubTLVLength uint8
	SubTLVs      []interface{}
}
