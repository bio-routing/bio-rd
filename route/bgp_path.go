package route

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/bio-routing/tflow2/convert"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route/api"
)

// BGPPath represents a set of BGP path attributes
type BGPPath struct {
	PathIdentifier    uint32
	NextHop           bnet.IP
	LocalPref         uint32
	ASPath            types.ASPath
	ASPathLen         uint16
	Origin            uint8
	MED               uint32
	EBGP              bool
	AtomicAggregate   bool
	Aggregator        *types.Aggregator
	BGPIdentifier     uint32
	Source            bnet.IP
	Communities       []uint32
	LargeCommunities  []types.LargeCommunity
	UnknownAttributes []types.UnknownPathAttribute
	OriginatorID      uint32
	ClusterList       []uint32
}

// ToProto converts BGPPath to proto BGPPath
func (b *BGPPath) ToProto() *api.BGPPath {
	if b == nil {
		return nil
	}

	a := &api.BGPPath{
		PathIdentifier:    b.PathIdentifier,
		NextHop:           b.NextHop.ToProto(),
		LocalPref:         b.LocalPref,
		AsPath:            b.ASPath.ToProto(),
		Origin:            uint32(b.Origin),
		Med:               b.MED,
		Ebgp:              b.EBGP,
		BgpIdentifier:     b.BGPIdentifier,
		Source:            b.Source.ToProto(),
		Communities:       make([]uint32, len(b.Communities)),
		LargeCommunities:  make([]*api.LargeCommunity, len(b.LargeCommunities)),
		UnknownAttributes: make([]*api.UnknownPathAttribute, len(b.UnknownAttributes)),
		OriginatorId:      b.OriginatorID,
		ClusterList:       make([]uint32, len(b.ClusterList)),
	}

	copy(a.Communities, b.Communities)
	copy(a.ClusterList, b.ClusterList)

	for i := range b.LargeCommunities {
		a.LargeCommunities[i] = b.LargeCommunities[i].ToProto()
	}

	for i := range b.UnknownAttributes {
		a.UnknownAttributes[i] = b.UnknownAttributes[i].ToProto()
	}

	return a
}

// BGPPathFromProtoBGPPath converts a proto BGPPath to BGPPath
func BGPPathFromProtoBGPPath(pb *api.BGPPath) *BGPPath {
	p := &BGPPath{
		PathIdentifier:    pb.PathIdentifier,
		NextHop:           bnet.IPFromProtoIP(*pb.NextHop),
		LocalPref:         pb.LocalPref,
		ASPath:            types.ASPathFromProtoASPath(pb.AsPath),
		Origin:            uint8(pb.Origin),
		MED:               pb.Med,
		EBGP:              pb.Ebgp,
		BGPIdentifier:     pb.BgpIdentifier,
		Source:            bnet.IPFromProtoIP(*pb.Source),
		Communities:       make([]uint32, len(pb.Communities)),
		LargeCommunities:  make([]types.LargeCommunity, len(pb.LargeCommunities)),
		UnknownAttributes: make([]types.UnknownPathAttribute, len(pb.UnknownAttributes)),
		OriginatorID:      pb.OriginatorId,
		ClusterList:       make([]uint32, len(pb.ClusterList)),
	}

	for i := range pb.Communities {
		p.Communities[i] = pb.Communities[i]
	}

	for i := range pb.LargeCommunities {
		p.LargeCommunities[i] = types.LargeCommunityFromProtoCommunity(pb.LargeCommunities[i])
	}

	for i := range pb.UnknownAttributes {
		p.UnknownAttributes[i] = types.UnknownPathAttributeFromProtoUnknownPathAttribute(pb.UnknownAttributes[i])
	}

	for i := range pb.ClusterList {
		p.ClusterList[i] = pb.ClusterList[i]
	}

	return p
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

	clusterListLen := uint16(0)
	if len(b.ClusterList) != 0 {
		clusterListLen += 3 + uint16(len(b.ClusterList)*4)
	}

	unknownAttributesLen := uint16(0)
	for _, unknownAttr := range b.UnknownAttributes {
		unknownAttributesLen += unknownAttr.WireLength()
	}

	originatorID := uint16(0)
	if b.OriginatorID != 0 {
		originatorID = 4
	}

	return communitiesLen + largeCommunitiesLen + 4*7 + 4 + originatorID + asPathLen + unknownAttributesLen
}

// ECMP determines if routes b and c are euqal in terms of ECMP
func (b *BGPPath) ECMP(c *BGPPath) bool {
	return b.LocalPref == c.LocalPref && b.ASPathLen == c.ASPathLen && b.MED == c.MED && b.Origin == c.Origin
}

// Equal checks if paths are equal
func (b *BGPPath) Equal(c *BGPPath) bool {
	if b.PathIdentifier != c.PathIdentifier {
		return false
	}

	return b.Select(c) == 0
}

// Select returns negative if b < c, 0 if paths are equal, positive if b > c
func (b *BGPPath) Select(c *BGPPath) int8 {
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

	// e) TODO: interior cost (hello IS-IS and OSPF)

	// f) + RFC4456 9. (Route Reflection)
	bgpIdentifierC := c.BGPIdentifier
	bgpIdentifierB := b.BGPIdentifier

	// IF an OriginatorID (set by an RR) is present, use this instead of Originator
	if c.OriginatorID != 0 {
		bgpIdentifierC = c.OriginatorID
	}

	if b.OriginatorID != 0 {
		bgpIdentifierB = b.OriginatorID
	}

	if bgpIdentifierC < bgpIdentifierB {
		return 1
	}

	if bgpIdentifierC > bgpIdentifierB {
		return -1
	}

	// Additionally check for the shorter ClusterList
	if len(c.ClusterList) < len(b.ClusterList) {
		return 1
	}

	if len(c.ClusterList) > len(b.ClusterList) {
		return -1
	}

	// g)
	if c.Source.Compare(b.Source) == -1 {
		return 1
	}

	if c.Source.Compare(b.Source) == 1 {
		return -1
	}

	if c.NextHop.Compare(b.NextHop) == -1 {
		return 1
	}

	if c.NextHop.Compare(b.NextHop) == 1 {
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

	if c.Source.Compare(b.Source) == -1 {
		return true
	}

	return false
}

// Print all known information about a route in logfile friendly format
func (b *BGPPath) String() string {
	buf := &strings.Builder{}

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

	fmt.Fprintf(buf, "Local Pref: %d, ", b.LocalPref)
	fmt.Fprintf(buf, "Origin: %s, ", origin)
	fmt.Fprintf(buf, "AS Path: %v, ", b.ASPath)
	fmt.Fprintf(buf, "BGP type: %s, ", bgpType)
	fmt.Fprintf(buf, "NEXT HOP: %s, ", b.NextHop)
	fmt.Fprintf(buf, "MED: %d, ", b.MED)
	fmt.Fprintf(buf, "Path ID: %d, ", b.PathIdentifier)
	fmt.Fprintf(buf, "Source: %s, ", b.Source)
	fmt.Fprintf(buf, "Communities: %v, ", b.Communities)
	fmt.Fprintf(buf, "LargeCommunities: %v", b.LargeCommunities)

	if b.OriginatorID != 0 {
		oid := convert.Uint32Byte(b.OriginatorID)
		fmt.Fprintf(buf, ", OriginatorID: %d.%d.%d.%d", oid[0], oid[1], oid[2], oid[3])
	}
	if b.ClusterList != nil {
		fmt.Fprintf(buf, ", ClusterList %s", b.ClusterListString())
	}

	return buf.String()
}

// Print all known information about a route in human readable form
func (b *BGPPath) Print() string {
	buf := &strings.Builder{}

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

	fmt.Fprintf(buf, "\t\tLocal Pref: %d\n", b.LocalPref)
	fmt.Fprintf(buf, "\t\tOrigin: %s\n", origin)
	fmt.Fprintf(buf, "\t\tAS Path: %v\n", b.ASPath)
	fmt.Fprintf(buf, "\t\tBGP type: %s\n", bgpType)
	fmt.Fprintf(buf, "\t\tNEXT HOP: %s\n", b.NextHop)
	fmt.Fprintf(buf, "\t\tMED: %d\n", b.MED)
	fmt.Fprintf(buf, "\t\tPath ID: %d\n", b.PathIdentifier)
	fmt.Fprintf(buf, "\t\tSource: %s\n", b.Source)
	fmt.Fprintf(buf, "\t\tCommunities: %v\n", b.Communities)
	fmt.Fprintf(buf, "\t\tLargeCommunities: %v\n", b.LargeCommunities)

	if b.OriginatorID != 0 {
		oid := convert.Uint32Byte(b.OriginatorID)
		fmt.Fprintf(buf, "\t\tOriginatorID: %d.%d.%d.%d\n", oid[0], oid[1], oid[2], oid[3])
	}
	if b.ClusterList != nil {
		fmt.Fprintf(buf, "\t\tClusterList %s\n", b.ClusterListString())
	}

	return buf.String()
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

	if b.ClusterList != nil {
		cp.ClusterList = make([]uint32, len(cp.ClusterList))
		copy(cp.ClusterList, b.ClusterList)
	}

	return &cp
}

// ComputeHash computes an hash over all attributes of the path
func (b *BGPPath) ComputeHash() string {
	s := fmt.Sprintf("%s\t%d\t%v\t%d\t%d\t%v\t%d\t%s\t%v\t%v\t%d\t%d\t%v",
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
		b.PathIdentifier,
		b.OriginatorID,
		b.ClusterList)

	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

// CommunitiesString returns the formated communities
func (b *BGPPath) CommunitiesString() string {
	str := &strings.Builder{}

	for i, com := range b.Communities {
		if i > 0 {
			str.WriteByte(' ')
		}
		str.WriteString(types.CommunityStringForUint32(com))
	}

	return str.String()
}

// ClusterListString returns the formated ClusterList
func (b *BGPPath) ClusterListString() string {
	str := &strings.Builder{}

	for i, cid := range b.ClusterList {
		if i > 0 {
			str.WriteByte(' ')
		}
		octes := convert.Uint32Byte(cid)

		fmt.Fprintf(str, "%d.%d.%d.%d", octes[0], octes[1], octes[2], octes[3])
	}

	return str.String()
}

// LargeCommunitiesString returns the formated communities
func (b *BGPPath) LargeCommunitiesString() string {
	str := &strings.Builder{}

	for i, com := range b.LargeCommunities {
		if i > 0 {
			str.WriteByte(' ')
		}
		str.WriteString(com.String())
	}

	return str.String()
}
