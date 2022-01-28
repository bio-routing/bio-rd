package packet

import (
	"bytes"

	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"
)

type Label uint32

func (l Label) serialize(buf *bytes.Buffer, bottomOfStack bool) {
	x := convert.Uint32Byte(uint32(l << 12))

	if bottomOfStack {
		x[2] |= 0x01
	}

	buf.Write(x[0:3])
}

func (l Label) isBottomOfStack() bool {
	return l&0x01 == 0x01
}

func decodeLabel(buf *bytes.Buffer) (Label, error) {
	label := make([]byte, BytesPerLabel)
	_, err := buf.Read(label)
	if err != nil {
		return Label(0), errors.Wrap(err, "read failed")
	}

	return Label(convert.Uint32b([]byte{0, label[0], label[1], label[2]})), nil
}
