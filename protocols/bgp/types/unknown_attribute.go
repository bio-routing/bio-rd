package types

// UnknownPathAttribute represents an unknown path attribute BIO does not support
type UnknownPathAttribute struct {
	Optional   bool
	Transitive bool
	Partial    bool
	TypeCode   uint8
	Value      []byte
}

// WireLength returns the number of bytes the attribute need on the wire
func (u *UnknownPathAttribute) WireLength() uint16 {
	length := uint16(len(u.Value))
	if length > 255 {
		length++ // Extended length
	}
	return length + 3
}
