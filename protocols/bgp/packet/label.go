package packet

import (
	"bytes"

	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"
)

const bottomOfStackBit = 0x01
const lengthEXPAndBottomOfStack = 0x04

type LabelStackEntry uint32

func (l LabelStackEntry) serialize(buf *bytes.Buffer, bottomOfStack bool) {
	x := convert.Uint32Byte(uint32(l << 12))

	if bottomOfStack {
		x[2] |= bottomOfStackBit
	}

	buf.Write(x[0:3])
}

func (l LabelStackEntry) isBottomOfStack() bool {
	return l&bottomOfStackBit == bottomOfStackBit
}

// GetLabel gets the label
func (l LabelStackEntry) GetLabel() uint32 {
	return uint32(l) >> lengthEXPAndBottomOfStack
}

func decodeLabel(buf *bytes.Buffer) (LabelStackEntry, error) {
	label := make([]byte, BytesPerLabel)
	_, err := buf.Read(label)
	if err != nil {
		return LabelStackEntry(0), errors.Wrap(err, "read failed")
	}

	return LabelStackEntry(convert.Uint32b([]byte{0, label[0], label[1], label[2]})), nil
}
