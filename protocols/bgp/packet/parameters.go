package packet

import (
	"bytes"

	"github.com/taktv6/tflow2/convert"
)

type Serializable interface {
	serialize(*bytes.Buffer)
}

type OptParam struct {
	Type   uint8
	Length uint8
	Value  Serializable
}

type Capability struct {
	Code   uint8
	Length uint8
	Value  Serializable
}

func (c Capability) serialize(buf *bytes.Buffer) {
	tmpBuf := bytes.NewBuffer(make([]byte, 0))
	c.Value.serialize(tmpBuf)
	payload := tmpBuf.Bytes()

	buf.WriteByte(c.Code)
	buf.WriteByte(uint8(len(payload)))
	buf.Write(payload)
}

type AddPathCapability struct {
	AFI         uint16
	SAFI        uint8
	SendReceive uint8
}

func (a AddPathCapability) serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint16Byte(a.AFI))
	buf.WriteByte(a.SAFI)
	buf.WriteByte(a.SendReceive)
}
