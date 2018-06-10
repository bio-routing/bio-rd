package packet

import (
	"bytes"
	"fmt"
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

const L2HelloMinSize = 18

func decodeISISHello(buf *bytes.Buffer) (*L2Hello, error) {
	pdu := &L2Hello{}
	fields := []interface{}{
		&pdu.CircuitType,
		&pdu.SystemID,
		&pdu.HoldingTimer,
		&pdu.PDULength,
		&pdu.Priority,
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
