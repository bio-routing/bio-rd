package ethernet

import "bytes"

// LLC is a Logical Link Control header
type LLC struct {
	DSAP         uint8
	SSAP         uint8
	ControlField uint8
}

func (l LLC) serialize(buf *bytes.Buffer) {
	buf.WriteByte(l.DSAP)
	buf.WriteByte(l.SSAP)
	buf.WriteByte(l.ControlField)
}
