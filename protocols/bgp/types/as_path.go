package types

import (
	"fmt"

	"github.com/bio-routing/bio-rd/route/api"
)

// ASPath Segment Types
const (
	// ASSet is the AS Path type used to indicate an AS Set (RFC4271)
	ASSet = 1

	// ASSequence is tha AS Path type used to indicate an AS Sequence (RFC4271)
	ASSequence = 2

	// MaxASNsSegment is the maximum number of ASNs in an AS segment
	MaxASNsSegment = 255
)

// ASPath represents an AS Path (RFC4271)
type ASPath []ASPathSegment

// ASPathSegment represents an AS Path Segment (RFC4271)
type ASPathSegment struct {
	Type uint8
	ASNs []uint32
}

// ToProto converts ASPath to proto ASPath
func (pa ASPath) ToProto() []*api.ASPathSegment {
	ret := make([]*api.ASPathSegment, len(pa))
	for i := range pa {
		ret[i] = &api.ASPathSegment{
			Asns: make([]uint32, len(pa[i].ASNs)),
		}

		if pa[i].Type == ASSequence {
			ret[i].AsSequence = true
		}

		copy(ret[i].Asns, pa[i].ASNs)
	}

	return ret
}

// ASPathFromProtoASPath converts an proto ASPath to ASPath
func ASPathFromProtoASPath(segments []*api.ASPathSegment) ASPath {
	asPath := make(ASPath, len(segments))

	for i := range segments {
		s := ASPathSegment{
			Type: ASSet,
			ASNs: make([]uint32, len(segments[i].Asns)),
		}

		if segments[i].AsSequence {
			s.Type = ASSequence
		}

		for j := range segments[i].Asns {
			s.ASNs[j] = segments[i].Asns[j]
		}

		asPath[i] = s
	}

	return asPath
}

// String converts an ASPath to it's human redable representation
func (pa ASPath) String() (ret string) {
	for _, p := range pa {
		if p.Type == ASSet {
			ret += " ("
		}
		n := len(p.ASNs)
		for i, asn := range p.ASNs {
			if i < n-1 {
				ret += fmt.Sprintf("%d ", asn)
				continue
			}
			ret += fmt.Sprintf("%d", asn)
		}
		if p.Type == ASSet {
			ret += ")"
		}
	}

	return
}

// Length returns the AS path length as used by path selection
func (pa ASPath) Length() (ret uint16) {
	for _, p := range pa {
		if p.Type == ASSet {
			ret++
			continue
		}
		ret += uint16(len(p.ASNs))
	}

	return
}
