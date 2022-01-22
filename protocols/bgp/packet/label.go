package packet

import (
	"bytes"

	"github.com/bio-routing/tflow2/convert"
)

type Label uint32

func (l Label) serialize(buf *bytes.Buffer, bottomOfStack bool) {
	x := convert.Uint32Byte(uint32(l << 12))

	if bottomOfStack {
		x[2] |= 0x01
	}

	buf.Write(x[0:3])
}
