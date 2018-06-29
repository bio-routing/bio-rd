package route

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/taktv6/tflow2/convert"
)

// BGPPath represents a set of BGP path attributes
type BGPPath struct {
	PathIdentifier    uint32
	NextHop           uint32
	LocalPref         uint32
	ASPath            types.ASPath
	ASPathLen         uint16
	Origin            uint8
	MED               uint32
	EBGP              bool
	AtomicAggregate   bool
	Aggregator        types.Aggregator
	BGPIdentifier     uint32
	Source            uint32
	Communities       []uint32
	LargeCommunities  []types.LargeCommunity
	UnknownAttributes []types.UnknownPathAttribute
}

// Length get's the length of serialized path
func (b *BGPPath) Length() uint16 {
	asPathLen := uint16(3)
	for _, segment := range b.ASPath {
		asPathLen++
		asPathLen += uint16(4 * len(segment.ASNs))
	}

	communitiesLen := uint16(0)
	if len(b.Communities) != 0 {
		communitiesLen += 3 + uint16(len(b.Communities)*4)
	}

	largeCommunitiesLen := uint16(0)
	if len(b.LargeCommunities) != 0 {
		largeCommunitiesLen += 3 + uint16(len(b.LargeCommunities)*12)
	}

	unknownAttributesLen := uint16(0)
	for _, unknownAttr := range b.UnknownAttributes {
		unknownAttributesLen += unknownAttr.WireLength()
	}

	return communitiesLen + largeCommunitiesLen + 4*7 + 4 + asPathLen + unknownAttributesLen
}

// ECMP determines if routes b and c are euqal in terms of ECMP
func (b *BGPPath) ECMP(c *BGPPath) bool {
	return b.LocalPref == c.LocalPref && b.ASPathLen == c.ASPathLen && b.MED == c.MED && b.Origin == c.Origin
}

// Compare returns negative if b < c, 0 if paths are equal, positive if b > c
func (b *BGPPath) Compare(c *BGPPath) int8 {
	if c.LocalPref < b.LocalPref {
		return 1
	}

	if c.LocalPref > b.LocalPref {
		return -1
	}

	// 9.1.2.2.  Breaking Ties (Phase 2)

	// a)
	if c.ASPathLen > b.ASPathLen {
		return 1
	}

	if c.ASPathLen < b.ASPathLen {
		return -1
	}

	// b)
	if c.Origin > b.Origin {
		return 1
	}

	if c.Origin < b.Origin {
		return -1
	}

	// c)
	if c.MED > b.MED {
		return 1
	}

	if c.MED < b.MED {
		return -1
	}

	// d)
	if c.EBGP && !b.EBGP {
		return -1
	}

	if !c.EBGP && b.EBGP {
		return 1
	}

	// e) TODO: interiour cost (hello IS-IS and OSPF)

	// f)
	if c.BGPIdentifier < b.BGPIdentifier {
		return 1
	}

	if c.BGPIdentifier > b.BGPIdentifier {
		return -1
	}

	// g)
	if c.Source < b.Source {
		return 1
	}

	if c.Source > b.Source {
		return -1
	}

	if c.NextHop < b.NextHop {
		return 1
	}

	if c.NextHop > b.NextHop {
		return -1
	}

	return 0
}

func (b *BGPPath) betterECMP(c *BGPPath) bool {
	if c.LocalPref < b.LocalPref {
		return false
	}

	if c.LocalPref > b.LocalPref {
		return true
	}

	if c.ASPathLen > b.ASPathLen {
		return false
	}

	if c.ASPathLen < b.ASPathLen {
		return true
	}

	if c.Origin > b.Origin {
		return false
	}

	if c.Origin < b.Origin {
		return true
	}

	if c.MED > b.MED {
		return false
	}

	if c.MED < b.MED {
		return true
	}

	return false
}

func (b *BGPPath) better(c *BGPPath) bool {
	if b.betterECMP(c) {
		return true
	}

	if c.BGPIdentifier < b.BGPIdentifier {
		return true
	}

	if c.Source < b.Source {
		return true
	}

	return false
}

// Print all known information about a route in human readable form
func (b *BGPPath) Print() string {
	origin := ""
	switch b.Origin {
	case 0:
		origin = "Incomplete"
	case 1:
		origin = "EGP"
	case 2:
		origin = "IGP"
	}

	bgpType := "internal"
	if b.EBGP {
		bgpType = "external"
	}

	ret := fmt.Sprintf("\t\tLocal Pref: %d\n", b.LocalPref)
	ret += fmt.Sprintf("\t\tOrigin: %s\n", origin)
	ret += fmt.Sprintf("\t\tAS Path: %v\n", b.ASPath)
	ret += fmt.Sprintf("\t\tBGP type: %s\n", bgpType)
	nh := uint32To4Byte(b.NextHop)
	ret += fmt.Sprintf("\t\tNEXT HOP: %d.%d.%d.%d\n", nh[0], nh[1], nh[2], nh[3])
	ret += fmt.Sprintf("\t\tMED: %d\n", b.MED)
	ret += fmt.Sprintf("\t\tPath ID: %d\n", b.PathIdentifier)
	src := uint32To4Byte(b.Source)
	ret += fmt.Sprintf("\t\tSource: %d.%d.%d.%d\n", src[0], src[1], src[2], src[3])
	ret += fmt.Sprintf("\t\tCommunities: %v\n", b.Communities)
	ret += fmt.Sprintf("\t\tLargeCommunities: %v\n", b.LargeCommunities)

	return ret
}

// Prepend the given BGPPath with the given ASN given times
func (b *BGPPath) Prepend(asn uint32, times uint16) {
	if times == 0 {
		return
	}

	if len(b.ASPath) == 0 {
		b.insertNewASSequence()
	}

	first := b.ASPath[0]
	if first.Type == types.ASSet {
		b.insertNewASSequence()
	}

	for i := 0; i < int(times); i++ {
		if len(b.ASPath) == types.MaxASNsSegment {
			b.insertNewASSequence()
		}

		old := b.ASPath[0].ASNs
		asns := make([]uint32, len(old)+1)
		copy(asns[1:], old)
		asns[0] = asn
		b.ASPath[0].ASNs = asns
	}

	b.ASPathLen = b.ASPath.Length()
}

func (b *BGPPath) insertNewASSequence() {
	pa := make(types.ASPath, len(b.ASPath)+1)
	copy(pa[1:], b.ASPath)
	pa[0] = types.ASPathSegment{
		ASNs: make([]uint32, 0),
		Type: types.ASSequence,
	}

	b.ASPath = pa
}

// Copy creates a deep copy of a BGPPath
func (b *BGPPath) Copy() *BGPPath {
	if b == nil {
		return nil
	}

	cp := *b

	cp.ASPath = make(types.ASPath, len(cp.ASPath))
	copy(cp.ASPath, b.ASPath)

	cp.Communities = make([]uint32, len(cp.Communities))
	copy(cp.Communities, b.Communities)

	cp.LargeCommunities = make([]types.LargeCommunity, len(cp.LargeCommunities))
	copy(cp.LargeCommunities, b.LargeCommunities)

	return &cp
}

// ComputeHash computes an hash over all attributes of the path
func (b *BGPPath) ComputeHash() string {
	s := fmt.Sprintf("%d\t%d\t%v\t%d\t%d\t%v\t%d\t%d\t%v\t%v\t%d",
		b.NextHop,
		b.LocalPref,
		b.ASPath,
		b.Origin,
		b.MED,
		b.EBGP,
		b.BGPIdentifier,
		b.Source,
		b.Communities,
		b.LargeCommunities,
		b.PathIdentifier)

	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

// CommunitiesString returns the formated communities
func (b *BGPPath) CommunitiesString() string {
	str := ""
	for _, com := range b.Communities {
		str += types.CommunityStringForUint32(com) + " "
	}

	return strings.TrimRight(str, " ")
}

// LargeCommunitiesString returns the formated communities
func (b *BGPPath) LargeCommunitiesString() string {
	str := ""
	for _, com := range b.LargeCommunities {
		str += com.String() + " "
	}

	return strings.TrimRight(str, " ")
}

func uint32To4Byte(addr uint32) [4]byte {
	slice := convert.Uint32Byte(addr)
	ret := [4]byte{
		slice[0],
		slice[1],
		slice[2],
		slice[3],
	}
	return ret
}
