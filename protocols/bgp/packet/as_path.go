package packet

import "fmt"

type ASPath []ASPathSegment

type ASPathSegment struct {
	Type  uint8
	Count uint8
	ASNs  []uint32
}

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
