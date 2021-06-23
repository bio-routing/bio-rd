package net

import (
	"encoding/binary"
	"unsafe"

	"github.com/bio-routing/tflow2/convert"
)

var (
	localEndianness binary.ByteOrder
)

func init() {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		localEndianness = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		localEndianness = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}
}

// Htons converts a local short uint to an network order short uint
func Htons(x uint16) uint16 {
	if localEndianness == binary.BigEndian {
		return x
	}

	xp := unsafe.Pointer(&x)
	b := (*[2]byte)(xp)

	tmp := b[0]
	b[0] = b[1]
	b[1] = tmp

	return *(*uint16)(xp)
}

// BigEndianToLocal converts input from big endian to "local" endian
func BigEndianToLocal(input []byte) []byte {
	if localEndianness == binary.BigEndian {
		return input
	}

	return convert.Reverse(input)
}
