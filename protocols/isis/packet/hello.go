package packet

import (
	"bytes"
	"fmt"

	"github.com/taktv6/tflow2/convert"
)

type L2Hello struct {
	CircuitType  uint8
	SystemID     [6]byte
	HoldingTimer uint16
	PDULength    uint16
	Priority     uint8
	DesignatedIS [6]byte
	TLVs         []TLV
}

const (
	L2HelloMinSize = 18
	L2CircuitType  = 2
)

func decodeISISHello(buf *bytes.Buffer) (*L2Hello, error) {
	pdu := &L2Hello{}
	reserved := uint8(0)
	fields := []interface{}{
		&pdu.CircuitType,
		&pdu.SystemID,
		&pdu.HoldingTimer,
		&pdu.PDULength,
		&pdu.Priority,
		&reserved,
		&pdu.DesignatedIS,
	}

	err := decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	TLVs, err := readTLVs(buf, pdu.PDULength-L2HelloMinSize)
	if err != nil {
		return nil, fmt.Errorf("Unable to read TLVs: %v", err)
	}

	pdu.TLVs = TLVs
	return pdu, nil
}

func (h *L2Hello) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(h.CircuitType)
	buf.Write(h.SystemID[:])
	buf.Write(convert.Uint16Byte(h.HoldingTimer))
	buf.Write(convert.Uint16Byte(h.PDULength))
	buf.WriteByte(h.Priority)
	buf.WriteByte(0) // Reserved
	buf.Write(h.DesignatedIS[:])
}
