package route

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/taktv6/tflow2/convert"
)

// BGPPath represents a set of BGP path attributes
type BGPPath struct {
	PathIdentifier   uint32
	NextHop          uint32
	LocalPref        uint32
	ASPath           packet.ASPath
	ASPathLen        uint16
	Origin           uint8
	MED              uint32
	EBGP             bool
	BGPIdentifier    uint32
	Source           uint32
	Communities      []uint32
	LargeCommunities []packet.LargeCommunity
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
	if first.Type == packet.ASSet {
		b.insertNewASSequence()
	}

	for i := 0; i < int(times); i++ {
		if len(b.ASPath) == packet.MaxASNsSegment {
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

func (b *BGPPath) insertNewASSequence() packet.ASPath {
	pa := make(packet.ASPath, len(b.ASPath)+1)
	copy(pa[1:], b.ASPath)
	pa[0] = packet.ASPathSegment{
		ASNs:  make([]uint32, 0),
		Count: 0,
		Type:  packet.ASSequence,
	}

	return pa
}

func (p *BGPPath) Copy() *BGPPath {
	if p == nil {
		return nil
	}

	cp := *p
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
		str += packet.CommunityStringForUint32(com) + " "
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
