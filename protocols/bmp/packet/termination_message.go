package packet

import (
	"bytes"
	"fmt"
)

// TerminationMessage represents a termination message
type TerminationMessage struct {
	CommonHeader *CommonHeader
	TLVs         []*InformationTLV
}

// MsgType returns the type of this message
func (t *TerminationMessage) MsgType() uint8 {
	return t.CommonHeader.MsgType
}

func decodeTerminationMessage(buf *bytes.Buffer, ch *CommonHeader) (*TerminationMessage, error) {
	tm := &TerminationMessage{
		CommonHeader: ch,
		TLVs:         make([]*InformationTLV, 0, 2),
	}

	read := uint32(0)
	toRead := ch.MsgLength - CommonHeaderLen

	for read < toRead {
		tlv, err := decodeInformationTLV(buf)
		if err != nil {
			return nil, fmt.Errorf("unable to decode TLV: %w", err)
		}

		tm.TLVs = append(tm.TLVs, tlv)
		read += uint32(tlv.InformationLength) + MinInformationTLVLen
	}

	return tm, nil
}
