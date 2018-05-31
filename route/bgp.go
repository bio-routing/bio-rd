package route

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/taktv6/tflow2/convert"
)

// BGPPath represents a set of BGP path attributes
type BGPPath struct {
	PathIdentifier   uint32
	NextHop          uint32
	LocalPref        uint32
	ASPath           string
	ASPathLen        uint16
	Origin           uint8
	MED              uint32
	EBGP             bool
	BGPIdentifier    uint32
	Source           uint32
	Communities      string
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
	ret += fmt.Sprintf("\t\tAS Path: %s\n", b.ASPath)
	nh := uint32To4Byte(b.NextHop)
	ret += fmt.Sprintf("\t\tNEXT HOP: %d.%d.%d.%d\n", nh[0], nh[1], nh[2], nh[3])
	ret += fmt.Sprintf("\t\tMED: %d\n", b.MED)

	return ret
}

func (b *BGPPath) Prepend(asn uint32, times uint16) {
	if times == 0 {
		return
	}

	asnStr := strconv.FormatUint(uint64(asn), 10)

	path := make([]string, times+1)
	for i := 0; uint16(i) < times; i++ {
		path[i] = asnStr
	}
	path[times] = b.ASPath

	b.ASPath = strings.TrimSuffix(strings.Join(path, " "), " ")
	b.ASPathLen = b.ASPathLen + times
}

func (p *BGPPath) Copy() *BGPPath {
	if p == nil {
		return nil
	}

	cp := *p
	return &cp
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
