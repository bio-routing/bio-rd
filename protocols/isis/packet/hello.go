package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/taktv6/tflow2/convert"
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
	P2PHelloMinLen = 20
	ISISHeaderSize = 8
	L2CircuitType  = 2
)

// GetProtocolsSupportedTLV gets the protocols supported TLV
func (h *P2PHello) GetProtocolsSupportedTLV() *ProtocolsSupportedTLV {
	for _, tlv := range h.TLVs {
		if tlv.Type() != ProtocolsSupportedTLVType {
			continue
		}

		return tlv.(*ProtocolsSupportedTLV)
	}

	return nil
}

// GetAreaAddressesTLV gets the area addresses TLV
func (h *P2PHello) GetAreaAddressesTLV() *AreaAddressesTLV {
	for _, tlv := range h.TLVs {
		if tlv.Type() != AreaAddressesTLVType {
			continue
		}

		return tlv.(*AreaAddressesTLV)
	}

	return nil
}

// GetP2PAdjTLV gets the P2P Adjacency TLV from the P2P Hello
func (h *P2PHello) GetP2PAdjTLV() *P2PAdjacencyStateTLV {
	for _, tlv := range h.TLVs {
		if tlv.Type() != P2PAdjacencyStateTLVType {
			continue
		}

		return tlv.(*P2PAdjacencyStateTLV)
	}

	return nil
}

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
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	TLVs, err := readTLVs(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to read TLVs: %v", err)
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
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	TLVs, err := readTLVs(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to read TLVs: %v", err)
	}

	pdu.TLVs = TLVs
	return pdu, nil
}
