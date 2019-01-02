package packet

import (
	"bytes"

	"github.com/bio-routing/bio-rd/util/decoder"
	"github.com/bio-routing/tflow2/convert"
)

const (
	// CommonHeaderLen is the length of a common header
	CommonHeaderLen = 6
)

// CommonHeader represents a common header
type CommonHeader struct {
	Version   uint8
	MsgLength uint32
	MsgType   uint8
}

// Serialize serializes a common header
func (c *CommonHeader) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(c.Version)
	buf.Write(convert.Uint32Byte(c.MsgLength))
	buf.WriteByte(c.MsgType)
}

func decodeCommonHeader(buf *bytes.Buffer) (*CommonHeader, error) {
	ch := &CommonHeader{}
	fields := []interface{}{
		&ch.Version,
		&ch.MsgLength,
		&ch.MsgType,
	}

	err := decoder.Decode(buf, fields)
	if err != nil {
		return ch, err
	}

	return ch, nil
}
