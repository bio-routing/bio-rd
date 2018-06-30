package types

import "fmt"

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
