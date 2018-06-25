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
	LargeCommunities string
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

	if c.ASPathLen > b.ASPathLen {
		return 1
	}

	if c.ASPathLen < b.ASPathLen {
		return -1
	}

	if c.Origin > b.Origin {
		return 1
	}

	if c.Origin < b.Origin {
		return -1
	}

	if c.MED > b.MED {
		return 1
	}

	if c.MED < b.MED {
		return -1
	}

	if c.BGPIdentifier < b.BGPIdentifier {
		return 1
	}

	if c.BGPIdentifier > b.BGPIdentifier {
		return -1
	}

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
	ret := fmt.Sprintf("\t\tLocal Pref: %d\n", b.LocalPref)
	ret += fmt.Sprintf("\t\tOrigin: %s\n", origin)
	ret += fmt.Sprintf("\t\tAS Path: %s\n")
	nh := uint32To4Byte(b.NextHop)
	ret += fmt.Sprintf("\t\tNEXT HOP: %d.%d.%d.%d\n", nh[0], nh[1], nh[2], nh[3])
	ret += fmt.Sprintf("\t\tMED: %d\n", b.MED)
	ret += fmt.Sprintf("\t\tPath ID: %d\n", b.PathIdentifier)
	ret += fmt.Sprintf("\t\tSource: %d\n", b.Source)
	ret += fmt.Sprintf("\t\tCommunities: %s\n", b.Communities)
	ret += fmt.Sprintf("\t\tLargeCommunities: %s\n", b.LargeCommunities)

	return ret
}

func (b *BGPPath) Prepend(asn uint32, times uint16) {
	if times == 0 {
		return
	}

	first := b.ASPath[0]
	if len(b.ASPath) == 0 || first.Type == packet.ASSet {
		b.insertNewASSequence()
	}

	for i := 0; i < int(times); i++ {
		if len(b.ASPath) == packet.MaxASNsSegment {
			b.insertNewASSequence()
		}

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
	s := fmt.Sprintf("%d\t%d\t%s\t%d\t%d\t%v\t%d\t%d\t%s\t%s\t%d",
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

	r := strings.NewReader(s)
	h := sha256.New()
	r.WriteTo(h)

	return fmt.Sprintf("%x", h.Sum(nil))
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
