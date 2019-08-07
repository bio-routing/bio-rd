package types

import (
	"github.com/bio-routing/bio-rd/route/api"
)

// UnknownPathAttribute represents an unknown path attribute BIO does not support
type UnknownPathAttribute struct {
	Optional   bool
	Transitive bool
	Partial    bool
	TypeCode   uint8
	Value      []byte
}

// Compare compares unknown attributes
func (u *UnknownPathAttribute) Compare(s *UnknownPathAttribute) bool {
	if u.Optional != s.Optional || u.Transitive != s.Transitive || u.Partial != s.Partial || u.TypeCode != s.TypeCode {
		return false
	}

	if len(u.Value) != len(s.Value) {
		return false
	}

	for i := range u.Value {
		if u.Value[i] != s.Value[i] {
			return false
		}
	}

	return true
}

// WireLength returns the number of bytes the attribute need on the wire
func (u *UnknownPathAttribute) WireLength() uint16 {
	length := uint16(len(u.Value))
	if length > 255 {
		length++ // Extended length
	}
	return length + 3
}

// ToProto converts UnknownPathAttribute to proto UnknownPathAttribute
func (u *UnknownPathAttribute) ToProto() *api.UnknownPathAttribute {
	a := &api.UnknownPathAttribute{
		Optional:   u.Optional,
		Transitive: u.Transitive,
		Partial:    u.Partial,
		TypeCode:   uint32(u.TypeCode),
		Value:      make([]byte, len(u.Value)),
	}

	copy(a.Value, u.Value)
	return a
}

// UnknownPathAttributeFromProtoUnknownPathAttribute convers an proto UnknownPathAttribute to UnknownPathAttribute
func UnknownPathAttributeFromProtoUnknownPathAttribute(x *api.UnknownPathAttribute) UnknownPathAttribute {
	return UnknownPathAttribute{
		Optional:   x.Optional,
		Transitive: x.Transitive,
		Partial:    x.Partial,
		TypeCode:   uint8(x.TypeCode),
		Value:      x.Value,
	}
}
