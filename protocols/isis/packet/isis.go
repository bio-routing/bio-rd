package packet

import "bytes"

type isisHeader struct {
	ProtoDiscriminator  uint8
	LengthIndicator     uint8
	ProtocolIDExtension uint8
	IDLength            uint8
	PDUType             uint8
	Version             uint8
	MaxAreaAddresses    uint8
}

func (h *isisHeader) serialize(buf *bytes.Buffer) {
	buf.WriteByte(h.ProtoDiscriminator)
	buf.WriteByte(h.LengthIndicator)
	buf.WriteByte(h.ProtocolIDExtension)
	buf.WriteByte(h.IDLength)
	buf.WriteByte(h.PDUType)
	buf.WriteByte(h.Version)
	buf.WriteByte(0) // Reserved
	buf.WriteByte(h.MaxAreaAddresses)
}

func DecodeHeader(buf *bytes.Buffer) (*isisHeader, error) {
	h := &isisHeader{}

	return h, nil
}
