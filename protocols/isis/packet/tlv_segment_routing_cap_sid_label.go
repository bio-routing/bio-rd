package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
)

const SRCapSidLabelTLVType = 1

// SRSidLabelTLV represents a Segment Routing SID/Label TLV as described in
// RFC 8665 Section 3.3.1
// +------+--------+-------+-------+-------+
// | Type | Length |  SID/Label (3 octets) |
// +------+--------+-------+-------+-------+

type SRCapSidLabelTLV struct {
	TLVType   uint8
	TLVLength uint8
	Label     [3]byte
}

const SRCapSidLabelTLVLength = 3

func readSRSidLabelTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*SRCapSidLabelTLV, error) {
	pdu := &SRCapSidLabelTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}
	fields := []any{
		&pdu.Label,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	return pdu, nil
}

func (s SRCapSidLabelTLV) Copy() TLV {
	ret := s
	return &ret
}

// Type gets the type of the TLV
func (s SRCapSidLabelTLV) Type() uint8 {
	return s.TLVType
}

// Length gets the length of the TLV
func (s SRCapSidLabelTLV) Length() uint8 {
	return s.TLVLength
}

// Value returns the TLV itself
func (s *SRCapSidLabelTLV) Value() any {
	return s
}

// Serialize serializes an SRSidLabelTLV
func (s SRCapSidLabelTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(s.TLVType)
	buf.WriteByte(s.TLVLength)
	buf.Write(s.Label[:])
}
