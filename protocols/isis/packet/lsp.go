package packet

import (
	"bytes"

	"github.com/FMNSSun/libhash/fletcher"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/pkg/errors"
	"github.com/taktv6/tflow2/convert"
)

const (
	LSPIDLen    = 8
	LSPDUMinLen = 19
)

// LSPID represents a Link State Packet ID
type LSPID struct {
	SystemID     types.SystemID
	PseudonodeID uint16
}

// Serialize serializes an LSPID
func (l *LSPID) Serialize(buf *bytes.Buffer) {
	buf.Write(l.SystemID[:])
	buf.Write(convert.Uint16Byte(l.PseudonodeID))
}

// LSPDU represents a link state PDU
type LSPDU struct {
	Length            uint16
	RemainingLifetime uint16
	LSPID             LSPID
	SequenceNumber    uint32
	Checksum          uint16
	TypeBlock         uint8
	TLVs              []TLV
}

// SetChecksum sets the checksum of an LSPDU
func (l *LSPDU) SetChecksum() {
	buf := bytes.NewBuffer(nil)
	l.Serialize(buf)

	x := fletcher.New16()
	x.Write(buf.Bytes())
	csum := x.Sum([]byte{})

	l.Checksum = uint16(csum[0])*256 + uint16(csum[1])
}

// Serialize serializes a linke state PDU
func (l *LSPDU) Serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint16Byte(l.Length))
	buf.Write(convert.Uint16Byte(l.RemainingLifetime))
	l.LSPID.Serialize(buf)
	buf.Write(convert.Uint32Byte(l.SequenceNumber))
	buf.Write(convert.Uint16Byte(l.Checksum))
	buf.WriteByte(l.TypeBlock)

	for _, TLV := range l.TLVs {
		TLV.Serialize(buf)
	}
}

// DecodeLSPDU decodes an LSPDU
func DecodeLSPDU(buf *bytes.Buffer) (*LSPDU, error) {
	pdu := &LSPDU{}

	fields := []interface{}{
		&pdu.Length,
		&pdu.RemainingLifetime,
		&pdu.LSPID.SystemID,
		&pdu.LSPID.PseudonodeID,
		&pdu.SequenceNumber,
		&pdu.Checksum,
		&pdu.TypeBlock,
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
