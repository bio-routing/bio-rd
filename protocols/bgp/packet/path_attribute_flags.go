package packet

import (
	"bytes"

	"github.com/bio-routing/bio-rd/util/decode"
)

func decodePathAttrFlags(buf *bytes.Buffer, pa *PathAttribute) error {
	flags := uint8(0)

	err := decode.DecodeUint8(buf, &flags)
	if err != nil {
		return err
	}

	pa.Optional = isOptional(flags)
	pa.Transitive = isTransitive(flags)
	pa.Partial = isPartial(flags)
	pa.ExtendedLength = isExtendedLength(flags)

	return nil
}

func isOptional(x uint8) bool {
	return x&128 == 128
}

func isTransitive(x uint8) bool {
	return x&64 == 64
}

func isPartial(x uint8) bool {
	return x&32 == 32
}

func isExtendedLength(x uint8) bool {
	return x&16 == 16
}

func setOptional(x uint8) uint8 {
	return x | 128
}

func setTransitive(x uint8) uint8 {
	return x | 64
}

func setPartial(x uint8) uint8 {
	return x | 32
}

func setExtendedLength(x uint8) uint8 {
	return x | 16
}
