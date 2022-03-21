package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/tflow2/convert"
)

const bottomOfStackBit = 0x01
const lengthEXPAndBottomOfStack = 0x04

type LabelStackEntry uint32

// NewLabelStackEntry creates a new label stack entry
func NewLabelStackEntry(labelValue uint32) LabelStackEntry {
	return LabelStackEntry(labelValue << lengthEXPAndBottomOfStack)
}

func (l LabelStackEntry) serialize(buf *bytes.Buffer, bottomOfStack bool) {
	x := convert.Uint32Byte(uint32(l))

	if bottomOfStack {
		x[3] |= bottomOfStackBit
	}

	buf.Write(x[1:4])
}

func (l LabelStackEntry) isBottomOfStack() bool {
	return l&bottomOfStackBit == bottomOfStackBit
}

// GetLabel gets the label
func (l LabelStackEntry) GetLabel() uint32 {
	return uint32(l) >> lengthEXPAndBottomOfStack
}

func decodeLabelStackEntry(buf *bytes.Buffer) (LabelStackEntry, error) {
	label := make([]byte, BytesPerLabel)
	_, err := buf.Read(label)
	if err != nil {
		return LabelStackEntry(0), fmt.Errorf("read failed: %w", err)
	}

	return LabelStackEntry(convert.Uint32b([]byte{0, label[0], label[1], label[2]})), nil
}
