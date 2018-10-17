package packet

import (
	"bytes"
	"fmt"
)

// InitiationMessage represents an initiation message
type InitiationMessage struct {
	CommonHeader *CommonHeader
	TLVs         []*InformationTLV
}

// MsgType returns the type of this message
func (im *InitiationMessage) MsgType() uint8 {
	return im.CommonHeader.MsgType
}

func decodeInitiationMessage(buf *bytes.Buffer, ch *CommonHeader) (Msg, error) {
	im := &InitiationMessage{
		CommonHeader: ch,
		TLVs:         make([]*InformationTLV, 0, 2),
	}

	read := uint32(0)
	toRead := ch.MsgLength - CommonHeaderLen

	for read < toRead {
		tlv, err := decodeInformationTLV(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode TLV: %v", err)
		}

		im.TLVs = append(im.TLVs, tlv)
		read += uint32(tlv.InformationLength) + MinInformationTLVLen
	}

	return im, nil
}
