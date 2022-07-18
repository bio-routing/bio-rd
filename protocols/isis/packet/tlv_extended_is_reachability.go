package packet

import (
	"bytes"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/tflow2/convert"
)

const (
	// ExtendedISReachabilityType is the type value of an Extended IS Reachability TLV
	ExtendedISReachabilityType = 22

	// LinkLocalRemoteIdentifiersSubTLVType is the type value of an Link Local/Remote Indentifiers Sub TLV
	LinkLocalRemoteIdentifiersSubTLVType = 4

	// IPv4InterfaceAddressSubTLVType is the type value of an IPv4 interface address sub TLV
	IPv4InterfaceAddressSubTLVType = 6

	// IPv4NeighborAddressSubTLVType is the type value of an IPv4 neighbor address sub TLV
	IPv4NeighborAddressSubTLVType = 8

	//ExtendedISReachabilityNeighborMinLen is the IS Rech. Neigh. min length
	ExtendedISReachabilityNeighborMinLen = 11
)

// ExtendedISReachabilityTLV is an Extended IS Reachability TLV
type ExtendedISReachabilityTLV struct {
	TLVType   uint8
	TLVLength uint8
	Neighbors []*ExtendedISReachabilityNeighbor
}

func (e *ExtendedISReachabilityTLV) Copy() TLV {
	ret := *e
	ret.Neighbors = make([]*ExtendedISReachabilityNeighbor, 0, len(e.Neighbors))
	for _, n := range e.Neighbors {
		ret.Neighbors = append(ret.Neighbors, n.Copy())
	}

	return &ret
}

// Type gets the type of the TLV
func (e *ExtendedISReachabilityTLV) Type() uint8 {
	return e.TLVType
}

// Length gets the length of the TLV
func (e *ExtendedISReachabilityTLV) Length() uint8 {
	return e.TLVLength
}

// Value returns the TLV itself
func (e *ExtendedISReachabilityTLV) Value() interface{} {
	return e
}

// AddNeighbor adds a neighbor to the extended IS Reach. TLV
func (e *ExtendedISReachabilityTLV) AddNeighbor(n *ExtendedISReachabilityNeighbor) {
	e.TLVLength += ExtendedISReachabilityNeighborMinLen + n.SubTLVLength
	e.Neighbors = append(e.Neighbors, n)
}

// Serialize serializes an ExtendedISReachabilityTLV
func (e *ExtendedISReachabilityTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(e.TLVType)
	buf.WriteByte(e.TLVLength)
	for i := range e.Neighbors {
		e.Neighbors[i].Serialize(buf)
	}
}

// ExtendedISReachabilityNeighbor is an extended IS Reachability Neighbor
type ExtendedISReachabilityNeighbor struct {
	NeighborID   types.SourceID
	Metric       uint32
	SubTLVLength uint8
	SubTLVs      []TLV
}

func (e *ExtendedISReachabilityNeighbor) Copy() *ExtendedISReachabilityNeighbor {
	ret := *e
	ret.SubTLVs = make([]TLV, 0, len(e.SubTLVs))
	for _, stlv := range e.SubTLVs {
		ret.SubTLVs = append(ret.SubTLVs, stlv.Copy())
	}
	return &ret
}

// Serialize serializes an ExtendedISReachabilityNeighbor
func (e *ExtendedISReachabilityNeighbor) Serialize(buf *bytes.Buffer) {
	buf.Write(e.NeighborID.Serialize())
	buf.Write(convert.Uint32Byte(e.Metric)[1:])
	buf.WriteByte(e.SubTLVLength)
	for i := range e.SubTLVs {
		e.SubTLVs[i].Serialize(buf)
	}
}

// NewExtendedISReachabilityNeighbor creates a new ExtendedISReachabilityNeighbor
func NewExtendedISReachabilityNeighbor(neighborID types.SourceID, metric uint32) *ExtendedISReachabilityNeighbor {
	return &ExtendedISReachabilityNeighbor{
		NeighborID: neighborID,
		Metric:     metric,
		SubTLVs:    make([]TLV, 0),
	}
}

// NewExtendedISReachabilityTLV creates a new Extended IS Reachability TLV
func NewExtendedISReachabilityTLV() *ExtendedISReachabilityTLV {
	e := &ExtendedISReachabilityTLV{
		TLVType:   ExtendedISReachabilityType,
		TLVLength: 0,
		Neighbors: make([]*ExtendedISReachabilityNeighbor, 0),
	}

	return e
}

// AddSubTLV adds a sub TLV to the ExtendedISReachabilityNeighbor
func (e *ExtendedISReachabilityNeighbor) AddSubTLV(tlv TLV) {
	e.SubTLVLength += tlv.Length() + 2
	e.SubTLVs = append(e.SubTLVs, tlv)
}

// LinkLocalRemoteIdentifiersSubTLV is an Link Local/Remote Identifiers Sub TLV
type LinkLocalRemoteIdentifiersSubTLV struct {
	TLVType   uint8
	TLVLength uint8
	Local     uint32
	Remote    uint32
}

func (l *LinkLocalRemoteIdentifiersSubTLV) Copy() TLV {
	ret := *l
	return &ret
}

// Type gets the type of the TLV
func (l *LinkLocalRemoteIdentifiersSubTLV) Type() uint8 {
	return l.TLVType
}

// Length gets the length of the TLV
func (l *LinkLocalRemoteIdentifiersSubTLV) Length() uint8 {
	return l.TLVLength
}

// Value returns the TLV itself
func (l *LinkLocalRemoteIdentifiersSubTLV) Value() interface{} {
	return l
}

// Serialize serializes an IPv4 address sub TLV
func (l *LinkLocalRemoteIdentifiersSubTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(l.TLVType)
	buf.WriteByte(l.TLVLength)
	buf.Write(convert.Uint32Byte(l.Local))
	buf.Write(convert.Uint32Byte(l.Remote))
}

// NewLinkLocalRemoteIdentifiersSubTLV creates a new LinkLocalRemoteIdentifiersSubTLV
func NewLinkLocalRemoteIdentifiersSubTLV(local uint32, remote uint32) *LinkLocalRemoteIdentifiersSubTLV {
	return &LinkLocalRemoteIdentifiersSubTLV{
		TLVType:   LinkLocalRemoteIdentifiersSubTLVType,
		TLVLength: 8,
		Local:     local,
		Remote:    remote,
	}
}

// IPv4AddressSubTLV is an IPv4 Address Sub TLV (used for both interface and neighbor)
type IPv4AddressSubTLV struct {
	TLVType   uint8
	TLVLength uint8
	Address   uint32
}

func (s *IPv4AddressSubTLV) Copy() TLV {
	ret := *s
	return &ret
}

// Type gets the type of the TLV
func (s *IPv4AddressSubTLV) Type() uint8 {
	return s.TLVType
}

// Length gets the length of the TLV
func (s *IPv4AddressSubTLV) Length() uint8 {
	return s.TLVLength
}

// Value returns the TLV itself
func (s *IPv4AddressSubTLV) Value() interface{} {
	return s
}

// Serialize serializes an IPv4 address sub TLV
func (s *IPv4AddressSubTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(s.TLVType)
	buf.WriteByte(s.TLVLength)
	buf.Write(convert.Uint32Byte(s.Address))
}

// NewIPv4InterfaceAddressSubTLV creates a new IPv4 Interface Address Sub TLV
func NewIPv4InterfaceAddressSubTLV(addr uint32) *IPv4AddressSubTLV {
	return newIPv4AddressSubTLV(IPv4InterfaceAddressSubTLVType, addr)
}

// NewIPv4NeighborAddressSubTLV creates a new IPv4 Neighbor Address Sub TLV
func NewIPv4NeighborAddressSubTLV(addr uint32) *IPv4AddressSubTLV {
	return newIPv4AddressSubTLV(IPv4NeighborAddressSubTLVType, addr)
}

func newIPv4AddressSubTLV(tlvType uint8, addr uint32) *IPv4AddressSubTLV {
	tlv := &IPv4AddressSubTLV{
		TLVType:   tlvType,
		TLVLength: 4,
		Address:   addr,
	}

	return tlv
}
