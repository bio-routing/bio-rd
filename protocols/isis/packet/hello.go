package packet

import (
	"bytes"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/pkg/errors"
	"github.com/bio-routing/tflow2/convert"
)

// L2Hello represents a broadcast L2 hello
type L2Hello struct {
	CircuitType  uint8
	SystemID     [6]byte
	HoldingTimer uint16
	PDULength    uint16
	Priority     uint8
	DesignatedIS [6]byte
	TLVs         []TLV
}

// P2PHello represents a Point to Point Hello
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

// Serialize serializes a P2P Hello
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

// DecodeP2PHello decodes a P2P Hello
func DecodeP2PHello(buf *bytes.Buffer) (*P2PHello, error) {
	pdu := &P2PHello{}

	fields := []interface{}{
		&pdu.CircuitType,
		&pdu.SystemID,
		&pdu.HoldingTimer,
		&pdu.PDULength,
		&pdu.LocalCircuitID,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to decode fields")
	}

	TLVs, err := readTLVs(buf)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read TLVs")
	}

	pdu.TLVs = TLVs
	return pdu, nil
}

// DecodeL2Hello decodes an ISIS broadcast L2 hello
func DecodeL2Hello(buf *bytes.Buffer) (*L2Hello, error) {
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

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to decode fields")
	}

	TLVs, err := readTLVs(buf)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read TLVs")
	}

	pdu.TLVs = TLVs
	return pdu, nil
}
