package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
)

const SegmentRoutingCapabilityTLVType = 2

type SegmentRoutingCapabilityTLV struct {
	TLVType   uint8
	TLVLength uint8
	Flags     uint8
	Range     [3]byte
	SubTLVs   []TLV
}

const SegmentRoutingCapabilityTLVLength = 4

func readSegmentRoutingCapabilityTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*SegmentRoutingCapabilityTLV, error) {
	pdu := &SegmentRoutingCapabilityTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}
	fields := []any{
		&pdu.Flags,
		&pdu.Range,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	toRead := tlvLength - SegmentRoutingCapabilityTLVLength
	if toRead > 0 {
		tlvsBytes := make([]byte, toRead)
		_, err := buf.Read(tlvsBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read TLV bytes from buf: %w", err)
		}

		subTLVs, err := readSRSubTLVs(bytes.NewBuffer(tlvsBytes))
		if err != nil {
			return nil, fmt.Errorf("unable to decode sub TLVs: %w", err)
		}

		pdu.SubTLVs = subTLVs
	}

	return pdu, nil
}

func (s SegmentRoutingCapabilityTLV) Copy() TLV {
	ret := s
	ret.SubTLVs = copyTLVs(s.SubTLVs)
	return &ret
}

// Type gets the type of the TLV
func (s SegmentRoutingCapabilityTLV) Type() uint8 {
	return s.TLVType
}

// Length gets the length of the TLV
func (s SegmentRoutingCapabilityTLV) Length() uint8 {
	return s.TLVLength
}

// Value returns the TLV itself
func (s *SegmentRoutingCapabilityTLV) Value() any {
	return s
}

func (s SegmentRoutingCapabilityTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(s.TLVType)
	buf.WriteByte(s.TLVLength)
	buf.WriteByte(s.Flags)
	buf.Write(s.Range[:])
	for _, tlv := range s.SubTLVs {
		tlv.Serialize(buf)
	}
}

func readSRSubTLVs(buf *bytes.Buffer) ([]TLV, error) {
	var tlvs []TLV
	for buf.Len() > 0 {
		tlv, err := readSRSubTLV(buf)
		if err != nil {
			return nil, fmt.Errorf("unable to decode TLV: %v", err)
		}
		tlvs = append(tlvs, tlv)
	}

	return tlvs, nil
}

func readSRSubTLV(buf *bytes.Buffer) (TLV, error) {
	var err error
	tlvType := uint8(0)
	tlvLength := uint8(0)

	headFields := []any{
		&tlvType,
		&tlvLength,
	}

	err = decode.Decode(buf, headFields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	switch tlvType {
	case SRCapSidLabelTLVType:
		return readSRSidLabelTLV(buf, tlvType, tlvLength)
	default:
		return readUnknownTLV(buf, tlvType, tlvLength)
	}
}
