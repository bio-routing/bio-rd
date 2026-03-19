package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
)

type IPv6TERouterIDTLV struct {
	TLVType   uint8
	TLVLength uint8
	RouterID  [16]byte
}

const IPv6TERouterIDTLVType = 12
const IPv6TERouterIDTLVLength = 18

func readIPv6TERouterIDTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*IPv6TERouterIDTLV, error) {
	pdu := &IPv6TERouterIDTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}

	fields := []any{
		&pdu.RouterID,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	return pdu, nil
}

func (i IPv6TERouterIDTLV) Copy() TLV {
	ret := i
	return &ret
}

// Type gets the type of the TLV
func (i IPv6TERouterIDTLV) Type() uint8 {
	return i.TLVType
}

// Length gets the length of the TLV
func (i IPv6TERouterIDTLV) Length() uint8 {
	return i.TLVLength
}

// Value returns the TLV itself
func (i *IPv6TERouterIDTLV) Value() any {
	return i
}

func (i *IPv6TERouterIDTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(i.TLVType)
	buf.WriteByte(i.TLVLength)
	buf.Write(i.RouterID[:])
}
