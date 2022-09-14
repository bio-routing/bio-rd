package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/tflow2/convert"
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
			return nil, fmt.Errorf("unable to decode TLV: %w", err)
		}

		im.TLVs = append(im.TLVs, tlv)
		read += uint32(tlv.InformationLength) + MinInformationTLVLen
	}

	return im, nil
}

func (im *InitiationMessage) Serialize(buf *bytes.Buffer) {
	im.setSizes()
	im.CommonHeader.Serialize(buf)

	for _, tlv := range im.TLVs {
		buf.Write(convert.Uint16Byte(tlv.InformationType))
		buf.Write(convert.Uint16Byte(tlv.InformationLength))
		buf.Write(tlv.Information)
	}
}

func (im *InitiationMessage) setSizes() {
	im.CommonHeader.MsgLength = CommonHeaderLen
	for _, tlv := range im.TLVs {
		tlv.InformationLength = uint16(len(tlv.Information))
		im.CommonHeader.MsgLength += uint32(MinInformationTLVLen + tlv.InformationLength)
	}
}
