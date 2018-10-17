package packet

// ISReachabilityTLVType is the type value of an IS reachability TLV
const ISReachabilityTLVType = 2

// ISReachabilityTLV represents an IS reachability TLV
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
