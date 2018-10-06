package packet

import (
	"bytes"
	"fmt"
)

type ISISHeader struct {
	ProtoDiscriminator  uint8
	LengthIndicator     uint8
	ProtocolIDExtension uint8
	IDLength            uint8
	PDUType             uint8
	Version             uint8
	MaxAreaAddresses    uint8
}

func (h *ISISHeader) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(h.ProtoDiscriminator)
	buf.WriteByte(h.LengthIndicator)
	buf.WriteByte(h.ProtocolIDExtension)
	buf.WriteByte(h.IDLength)
	buf.WriteByte(h.PDUType)
	buf.WriteByte(h.Version)
	buf.WriteByte(0) // Reserved
	buf.WriteByte(h.MaxAreaAddresses)
}

func decodeHeader(buf *bytes.Buffer) (*ISISHeader, error) {
	h := &ISISHeader{}
	dsap := uint8(0)
	ssap := uint8(0)
	cf := uint8(0)
	reserved := uint8(0)

	fields := []interface{}{
		&dsap,
		&ssap,
		&cf,
		&h.ProtoDiscriminator,
		&h.LengthIndicator,
		&h.ProtocolIDExtension,
		&h.IDLength,
		&h.PDUType,
		&h.Version,
		&reserved,
		&h.MaxAreaAddresses,
	}

	err := decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	return h, nil
}
