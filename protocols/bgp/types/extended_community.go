package types

import (
	"fmt"

	"github.com/bio-routing/bio-rd/route/api"
)

// TODO: Add EC type mappings

type ExtendedCommunities []ExtendedCommunity

// ExtendedCommunity represents an extended community as
// described in RFC 4360.
type ExtendedCommunity struct {
	Type    uint8
	SubType uint8
	Value   []byte
}

// ToProto converts ExtendedCommunity to proto ExtendedCommunity
func (ec *ExtendedCommunity) ToProto() *api.ExtendedCommunity {
	return &api.ExtendedCommunity{
		Type:    uint32(ec.Type),
		Subtype: uint32(ec.SubType),
		Value:   ec.Value,
	}
}

// ExtendedCommunityFromProtoExtendedCommunity converts a proto ExtendedCommunity to ExtendedCommunity
func ExtendedCommunityFromProtoExtendedCommunity(aec *api.ExtendedCommunity) ExtendedCommunity {
	return ExtendedCommunity{
		Type:    uint8(aec.Type),
		SubType: uint8(aec.Subtype),
		Value:   aec.Value,
	}
}

// String transitions an extended community its (almost) human-readable representation
func (ec *ExtendedCommunity) String() string {
	if ec == nil {
		return ""
	}

	var subType uint8
	if ec.SubType == 0 {
		subType = 0xff
	} else {
		subType = ec.SubType
	}

	// TODO: translate types and value
	return fmt.Sprintf("Type: %d Subtype: %d Value: %s", ec.Type, subType, ec.Value)
}