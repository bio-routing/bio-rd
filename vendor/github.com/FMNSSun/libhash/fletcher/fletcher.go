package fletcher

import "encoding/binary"
import "hash"

type Fletcher16 struct {
	s0 uint16
	s1 uint16
}

func New16() hash.Hash {
	return &Fletcher16{s0: 0, s1: 0}
}

func (s *Fletcher16) Reset() {
	s.s0 = 0
	s.s1 = 0
}

func (s *Fletcher16) Size() int {
	return 2
}

func (s *Fletcher16) BlockSize() int {
	return 256
}

func (s *Fletcher16) Sum(in []byte) []byte {
	temp := []byte{0, 0}
	binary.BigEndian.PutUint16(temp, (s.s1<<8)|(s.s0))
	return append(in, temp...)
}

func (s *Fletcher16) Write(in []byte) (int, error) {
	for _, v := range in {
		s.s0 = (s.s0 + uint16(v)) % 255
		s.s1 = (s.s0 + s.s1) % 255
	}

	return len(in), nil
}
