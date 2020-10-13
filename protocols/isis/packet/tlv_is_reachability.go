package packet

import (
	"bytes"
	"unsafe"
)

// ISReachabilityTLVType is the type value of an IS reachability TLV
const ISReachabilityTLVType = 2
const defaultLegacyMetric = 63

// ISReachabilityTLV represents an IS reachability TLV
type ISReachabilityTLV struct {
	TLVType     uint8
	TLVLength   uint8
	VirtualFlag uint8
	Neighbors   []ISNeighbor
}

// ISNeighbor is a neighbor within an ISReachabilityTLV
type ISNeighbor struct {
	RIEDefaultMetric uint8
	SIEDelayMetric   uint8
	SIEExpenseMetric uint8
	SIEErrorMetric   uint8
	NeighborID       [7]byte
}

// NewISReachabilityTLV creates a new IS Reachability TLV (type 2)
func NewISReachabilityTLV(neighbors [][7]byte) *ISReachabilityTLV {
	isrtlv := &ISReachabilityTLV{
		TLVType:   ISReachabilityTLVType,
		TLVLength: uint8(len(neighbors))*uint8(unsafe.Sizeof(ISNeighbor{})) + 1,
		Neighbors: make([]ISNeighbor, 0, len(neighbors)),
	}

	for _, n := range neighbors {
		isrtlv.Neighbors = append(isrtlv.Neighbors, newISNeighbor(n))
	}

	return isrtlv
}

func newISNeighbor(neighborID [7]byte) ISNeighbor {
	return ISNeighbor{
		RIEDefaultMetric: defaultLegacyMetric,
		SIEDelayMetric:   defaultLegacyMetric,
		SIEExpenseMetric: defaultLegacyMetric,
		SIEErrorMetric:   defaultLegacyMetric,
		NeighborID:       neighborID,
	}
}

// Type gets the type of the TLV
func (isrtlv *ISReachabilityTLV) Type() uint8 {
	return isrtlv.TLVType
}

// Length gets the length of the TLV
func (isrtlv *ISReachabilityTLV) Length() uint8 {
	return isrtlv.TLVLength
}

// Value gets the TLV itself
func (isrtlv *ISReachabilityTLV) Value() interface{} {
	return isrtlv
}

// Serialize serializes an IS Reachability TLV
func (isrtlv *ISReachabilityTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(isrtlv.TLVType)
	buf.WriteByte(isrtlv.TLVLength)
	buf.WriteByte(0) // Virtual Flag

	for _, n := range isrtlv.Neighbors {
		buf.WriteByte(n.RIEDefaultMetric)
		buf.WriteByte(n.SIEDelayMetric)
		buf.WriteByte(n.SIEExpenseMetric)
		buf.WriteByte(n.SIEErrorMetric)
		buf.Write(n.NeighborID[:])
	}
}
