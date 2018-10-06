package packet

const ISReachabilityTLVType = 2

type ISReachabilityTLV struct {
	TLVType          uint8
	TLVLength        uint8
	VirtualFlag      uint8
	RIEDefaultMetric uint8
	SIEDelayMetric   uint8
	SIEExpenseMetric uint8
	SIEErrorMetric   uint8
	NeighborID       [7]byte
}
