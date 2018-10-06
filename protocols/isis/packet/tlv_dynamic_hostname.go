package packet

import (
	"bytes"
	"fmt"
)

const DynamicHostNameTLVType = 137

// DynamicHostNameTLV represents a dynamic Hostname TLV
type DynamicHostNameTLV struct {
	TLVType   uint8
	TLVLength uint8
	Hostname  []byte
}

// Type gets the type of the TLV
func (d *DynamicHostNameTLV) Type() uint8 {
	return d.TLVType
}

// Length gets the length of the TLV
func (d *DynamicHostNameTLV) Length() uint8 {
	return d.TLVLength
}

// Value returns the dynamic hostname
func (d *DynamicHostNameTLV) Value() interface{} {
	return d.Hostname
}

func readDynamicHostnameTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*DynamicHostNameTLV, error) {
	pdu := &DynamicHostNameTLV{
		TLVType:   tlvType,
		TLVLength: tlvLength,
		Hostname:  make([]byte, tlvLength),
	}

	fields := []interface{}{
		&pdu.Hostname,
	}

	err := decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	return pdu, nil
}

// Serialize serializes a dynamic hostname TLV
func (d *DynamicHostNameTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(d.TLVType)
	buf.WriteByte(d.TLVLength)
	buf.Write(d.Hostname)
}
