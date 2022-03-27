package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"
)

const (
	// ExtendedIPReachabilityTLVType is the type value of an Extended IP Reachability TLV
	ExtendedIPReachabilityTLVType = 135

	// ExtendedIPReachabilityLength is the length of an Extended IP Reachability excluding Sub TLVs
	ExtendedIPReachabilityLength = 9
)

// ExtendedIPReachabilityTLV is an Extended IP Reachability TLV
type ExtendedIPReachabilityTLV struct {
	TLVType                  uint8
	TLVLength                uint8
	ExtendedIPReachabilities []*ExtendedIPReachability
}

// Type gets the type of the TLV
func (e *ExtendedIPReachabilityTLV) Type() uint8 {
	return e.TLVType
}

// Length gets the length of the TLV
func (e *ExtendedIPReachabilityTLV) Length() uint8 {
	return e.TLVLength
}

// Value returns the TLV itself
func (e *ExtendedIPReachabilityTLV) Value() interface{} {
	return e
}

// Serialize serializes an ExtendedIPReachabilityTLV
func (e *ExtendedIPReachabilityTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(e.TLVType)
	buf.WriteByte(e.TLVLength)

	for i := range e.ExtendedIPReachabilities {
		e.ExtendedIPReachabilities[i].Serialize(buf)
	}
}

// NewExtendedIPReachabilityTLV creates a new ExtendedIPReachabilityTLV
func NewExtendedIPReachabilityTLV() *ExtendedIPReachabilityTLV {
	return &ExtendedIPReachabilityTLV{
		TLVType:                  ExtendedIPReachabilityTLVType,
		TLVLength:                0,
		ExtendedIPReachabilities: make([]*ExtendedIPReachability, 0),
	}
}

func readExtendedIPReachabilityTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*ExtendedIPReachabilityTLV, error) {
	pdu := NewExtendedIPReachabilityTLV()
	pdu.TLVLength = tlvLength

	toRead := tlvLength
	for toRead > 0 {
		extIPReach, err := readExtendedIPReachability(buf)
		if err != nil {
			return nil, fmt.Errorf("unable to reach extended IP reachability: %w", err)
		}

		toRead -= ExtendedIPReachabilityLength
		for i := range extIPReach.SubTLVs {
			toRead -= extIPReach.SubTLVs[i].Length()
		}

		pdu.ExtendedIPReachabilities = append(pdu.ExtendedIPReachabilities, extIPReach)
	}

	return pdu, nil
}

// ExtendedIPReachability is the Extended IP Reachability Part of an ExtendedIPReachabilityTLV
type ExtendedIPReachability struct {
	Metric         uint32
	UDSubBitPfxLen uint8
	Address        uint32
	SubTLVs        []TLV
}

// NewExtendedIPReachability creates a new ExtendedIPReachability
func NewExtendedIPReachability(metric uint32, pfxLen uint8, addr uint32) *ExtendedIPReachability {
	return &ExtendedIPReachability{
		Metric:         metric,
		UDSubBitPfxLen: pfxLen,
		Address:        addr,
	}
}

// AddExtendedIPReachability adds an extended IP reachability
func (e *ExtendedIPReachabilityTLV) AddExtendedIPReachability(eipr *ExtendedIPReachability) {
	e.ExtendedIPReachabilities = append(e.ExtendedIPReachabilities, eipr)
	e.TLVLength += ExtendedIPReachabilityLength
	// TODO: Add length of sub TLVs. They will be added as soon as we support for TE
}

// Serialize serializes an ExtendedIPReachability
func (e *ExtendedIPReachability) Serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint32Byte(e.Metric))
	buf.WriteByte(e.UDSubBitPfxLen)
	buf.Write(convert.Uint32Byte(e.Address))

	for i := range e.SubTLVs {
		e.SubTLVs[i].Serialize(buf)
	}
}

func (e *ExtendedIPReachability) hasSubTLVs() bool {
	return e.UDSubBitPfxLen&(uint8(1)<<6) == 64
}

// PfxLen returns the prefix length
func (e *ExtendedIPReachability) PfxLen() uint8 {
	return (e.UDSubBitPfxLen << 2) >> 2
}

func readExtendedIPReachability(buf *bytes.Buffer) (*ExtendedIPReachability, error) {
	e := &ExtendedIPReachability{}

	fields := []interface{}{
		&e.Metric,
		&e.UDSubBitPfxLen,
		&e.Address,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	if !e.hasSubTLVs() {
		return e, nil
	}

	subTLVsLen := uint8(0)
	err = decode.Decode(buf, []interface{}{&subTLVsLen})
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	toRead := subTLVsLen
	for toRead > 0 {
		// TODO: Read Sub TLVs
	}

	return e, nil
}
