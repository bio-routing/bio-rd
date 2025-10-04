package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"
)

const RouterCapabilityTLVType = 242

type RouterCapabilityTLV struct {
	TLVType   uint8
	TLVLength uint8
	RouterID  uint32
	Flags     uint8
	SubTLVs   []TLV
}

const RouterCapabilityTLVLen = 5

func readRouterCapabilityTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*RouterCapabilityTLV, error) {
	pdu := &RouterCapabilityTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}
	fields := []interface{}{
		&pdu.RouterID,
		&pdu.Flags,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	toRead := tlvLength - RouterCapabilityTLVLen
	fmt.Printf("toRead: %d\n", toRead)
	if toRead > 0 {
		tlvsBytes := make([]byte, toRead)
		_, err := buf.Read(tlvsBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to read TLV bytes from buf: %w", err)
		}

		subTLVs, err := readRouterCapSubTLVs(bytes.NewBuffer(tlvsBytes))
		if err != nil {
			return nil, fmt.Errorf("unable to decode sub TLVs: %w", err)
		}

		pdu.SubTLVs = subTLVs
	}

	return pdu, nil
}

func (r RouterCapabilityTLV) Copy() TLV {
	ret := r
	ret.SubTLVs = copyTLVs(r.SubTLVs)

	return &ret
}

// Type returns the type of the TLV
func (r RouterCapabilityTLV) Type() uint8 {
	return r.TLVType
}

// Length returns the length of the TLV
func (r RouterCapabilityTLV) Length() uint8 {
	return r.TLVLength
}

// Value returns the TLV itself
func (r RouterCapabilityTLV) Value() interface{} {
	return r
}

// Serialize serializes an WriteByte into a buffer
func (r RouterCapabilityTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(r.TLVType)
	buf.WriteByte(r.TLVLength)
	buf.Write(convert.Uint32Byte(r.RouterID))
	buf.WriteByte(r.Flags)

	for i := range r.SubTLVs {
		r.SubTLVs[i].Serialize(buf)
	}
}

func readRouterCapSubTLVs(buf *bytes.Buffer) ([]TLV, error) {
	TLVs := make([]TLV, 0)
	for buf.Len() > 0 {
		tlv, err := readRouterCapSubTLV(buf)
		if err != nil {
			return nil, fmt.Errorf("unable to read TLV: %w", err)
		}

		TLVs = append(TLVs, tlv)
	}

	return TLVs, nil
}

func readRouterCapSubTLV(buf *bytes.Buffer) (TLV, error) {
	tlvType := uint8(0)
	tlvLength := uint8(0)

	headFields := []any{
		&tlvType,
		&tlvLength,
	}

	err := decode.Decode(buf, headFields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	var tlv TLV
	switch tlvType {
	case SegmentRoutingCapabilityTLVType:
		tlv, err = readSegmentRoutingCapabilityTLV(buf, tlvType, tlvLength)
	case SegmentRoutingAlgorithmsTLVType:
		tlv, err = readSegmentRoutingAlgorithmsTLV(buf, tlvType, tlvLength)
	case NodeMaximumSIDDepthTLVType:
		tlv, err = readNodeMaximumSIDDepthTLV(buf, tlvType, tlvLength)
	case IPv6TERouterIDTLVType:
		tlv, err = readIPv6TERouterIDTLV(buf, tlvType, tlvLength)
	default:
		tlv, err = readUnknownTLV(buf, tlvType, tlvLength)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to read TLV (type %d): %v", tlvType, err)
	}

	return tlv, nil
}
