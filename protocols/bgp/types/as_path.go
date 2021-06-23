package types

import (
	"strconv"
	"strings"

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

// Compare compares two AS Paths
func (a *ASPath) Compare(b *ASPath) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(*a) != len(*b) {
		return false
	}

	for i := range *a {
		if !(*a)[i].Compare((*b)[i]) {
			return false
		}
	}

	return true
}

// GetFirstSequenceSegment gets the first sequence of an AS path
func (a *ASPath) GetFirstSequenceSegment() *ASPathSegment {
	for _, seg := range *a {
		if seg.Type == ASSequence {
			return &seg
		}
	}

	return nil
}

// GetLastSequenceSegment gets the last sequence of an AS path
func (a *ASPath) GetLastSequenceSegment() *ASPathSegment {
	for i := len(*a) - 1; i >= 0; i-- {
		if (*a)[i].Type == ASSequence {
			return &(*a)[i]
		}
	}

	return nil
}

// ASPathSegment represents an AS Path Segment (RFC4271)
type ASPathSegment struct {
	Type uint8
	ASNs []uint32
}

// GetFirstASN returns the first ASN of an AS path segment
func (s ASPathSegment) GetFirstASN() *uint32 {
	if len(s.ASNs) == 0 {
		return nil
	}

	ret := s.ASNs[0]
	return &ret
}

// GetLastASN returns the last ASN of an AS path segment
func (s ASPathSegment) GetLastASN() *uint32 {
	if len(s.ASNs) == 0 {
		return nil
	}

	ret := s.ASNs[len(s.ASNs)-1]
	return &ret
}

// Compare checks if ASPathSegments are the same
func (s ASPathSegment) Compare(t ASPathSegment) bool {
	if s.Type != t.Type {
		return false
	}

	if len(s.ASNs) != len(t.ASNs) {
		return false
	}

	for i := range s.ASNs {
		if s.ASNs[i] != t.ASNs[i] {
			return false
		}
	}

	return true
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
func ASPathFromProtoASPath(segments []*api.ASPathSegment) *ASPath {
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

	return &asPath
}

// String converts an ASPath to it's human redable representation
func (a *ASPath) String() string {
	if a == nil {
		return ""
	}

	parts := make([]string, 0)

	for _, p := range *a {
		if p.Type == ASSequence {
			for _, asn := range p.ASNs {
				parts = append(parts, strconv.Itoa(int(asn)))
			}

			continue
		}

		if p.Type == ASSet {
			setParts := make([]string, len(p.ASNs))
			for i, asn := range p.ASNs {
				setParts[i] = strconv.Itoa(int(asn))
			}
			parts = append(parts, "("+strings.Join(setParts, " ")+")")
		}
	}

	return strings.Join(parts, " ")
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
