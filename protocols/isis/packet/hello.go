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

type P2PHello struct {
	CircuitType    uint8
	SystemID       [6]byte
	HoldingTimer   uint16
	PDULength      uint16
	LocalCircuitID uint8
	TLVs           []TLV
}

const (
	L2HelloMinSize  = 18
	P2PHelloMinSize = 12
	L2CircuitType   = 2
)

func (h *P2PHello) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(h.CircuitType)
	buf.Write(h.SystemID[:])
	buf.Write(convert.Uint16Byte(h.HoldingTimer))
	buf.Write(convert.Uint16Byte(h.PDULength))
	buf.WriteByte(h.LocalCircuitID)

	for _, TLV := range h.TLVs {
		TLV.Serialize(buf)
	}
}

func decodeISISP2PHello(buf *bytes.Buffer) (*P2PHello, error) {
	pdu := &P2PHello{}

	fields := []interface{}{
		&pdu.CircuitType,
		&pdu.SystemID,
		&pdu.HoldingTimer,
		&pdu.PDULength,
		&pdu.LocalCircuitID,
	}

	err := decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	TLVs, err := readTLVs(buf, pdu.PDULength-P2PHelloMinSize)
	if err != nil {
		return nil, fmt.Errorf("Unable to read TLVs: %v", err)
	}

	pdu.TLVs = TLVs
	return pdu, nil
}

func decodeISISL2Hello(buf *bytes.Buffer) (*L2Hello, error) {
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
