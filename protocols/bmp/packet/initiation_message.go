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

// SetCommonHeader sets the common header
func (im *InitiationMessage) SetCommonHeader(ch *CommonHeader) {
	im.CommonHeader = ch
}

func decodeInitiationMessage(buf *bytes.Buffer, ch *CommonHeader) (Msg, error) {
	im := &InitiationMessage{
		TLVs: make([]*InformationTLV, 0, 2),
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
		fmt.Printf("read: %d\n", read)
	}

	return im, nil
}