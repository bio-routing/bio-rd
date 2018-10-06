package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
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
	SystemID       types.SystemID
	HoldingTimer   uint16
	PDULength      uint16
	LocalCircuitID uint8
	TLVs           []TLV
}

const (
	P2PHelloMinSize = 20
	ISISHeaderSize  = 8
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

	TLVs, err := readTLVs(buf)
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

	TLVs, err := readTLVs(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to read TLVs: %v", err)
	}

	pdu.TLVs = TLVs
	return pdu, nil
}
