package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
)

const SegmentRoutingAlgorithmsTLVType = 19
const SegmentRoutingAlgorithmsTLVLength = 0

type SegmentRoutingAlgorithmsTLV struct {
	TLVType    uint8
	TLVLength  uint8
	Algorithms []uint8
}

func readSegmentRoutingAlgorithmsTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*SegmentRoutingAlgorithmsTLV, error) {
	pdu := &SegmentRoutingAlgorithmsTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
	}
	fields := []any{
		&pdu.Algorithms,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	algos := make([]uint8, tlvLength)
	_, err = buf.Read(algos)
	if err != nil {
		return nil, fmt.Errorf("unable to read algorithms: %v", err)
	}

	pdu.Algorithms = algos
	return pdu, nil
}

func (s SegmentRoutingAlgorithmsTLV) Copy() TLV {
	ret := s
	ret.Algorithms = make([]uint8, len(s.Algorithms))
	copy(ret.Algorithms, s.Algorithms)
	return &ret
}

// Type gets the type of the TLV
func (s SegmentRoutingAlgorithmsTLV) Type() uint8 {
	return s.TLVType
}

// Length gets the length of the TLV
func (s SegmentRoutingAlgorithmsTLV) Length() uint8 {
	return s.TLVLength
}

// Value returns the TLV itself
func (s *SegmentRoutingAlgorithmsTLV) Value() any {
	return s
}

func (s *SegmentRoutingAlgorithmsTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(s.TLVType)
	buf.WriteByte(s.TLVLength)
	buf.Write(s.Algorithms)
}
