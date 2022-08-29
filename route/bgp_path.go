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
	BGPPathA          *BGPPathA
	ASPath            *types.ASPath
	ClusterList       *types.ClusterList
	Communities       *types.Communities
	LargeCommunities  *types.LargeCommunities
	UnknownAttributes []types.UnknownPathAttribute
	PathIdentifier    uint32
	ASPathLen         uint16
	BMPPostPolicy     bool // BMPPostPolicy fields is a hack used in BMP to differentiate between pre/post policy routes (L flag of the per peer header)
}

// BGPPathA represents cachable BGP path attributes
type BGPPathA struct {
	NextHop         *bnet.IP
	Source          *bnet.IP
	LocalPref       uint32
	MED             uint32
	BGPIdentifier   uint32
	OriginatorID    uint32
	Aggregator      *types.Aggregator
	EBGP            bool
	AtomicAggregate bool
	Origin          uint8
	OnlyToCustomer  uint32
}

// NewBGPPathA creates a new BGPPathA
func NewBGPPathA() *BGPPathA {
	defaultAddr := bnet.IPv4(0)
	return &BGPPathA{
		NextHop: &defaultAddr,
		Source:  &defaultAddr,
	}
}

func (b *BGPPathA) Dedup() *BGPPathA {
	return bgpC.get(b)
}

func (b *BGPPath) Dedup() *BGPPath {
	b.BGPPathA = b.BGPPathA.Dedup()
	return b
}

// ToProto converts BGPPath to proto BGPPath
func (b *BGPPath) ToProto() *api.BGPPath {
	if b == nil {
		return nil
	}

	a := &api.BGPPath{
		PathIdentifier:    b.PathIdentifier,
		UnknownAttributes: make([]*api.UnknownPathAttribute, len(b.UnknownAttributes)),
		BmpPostPolicy:     b.BMPPostPolicy,
	}

	if b.BGPPathA != nil {
		a.LocalPref = b.BGPPathA.LocalPref
		a.Origin = uint32(b.BGPPathA.Origin)
		a.Med = b.BGPPathA.MED
		a.Ebgp = b.BGPPathA.EBGP
		a.BgpIdentifier = b.BGPPathA.BGPIdentifier
		a.OriginatorId = b.BGPPathA.OriginatorID
		a.OnlyToCustomer = b.BGPPathA.OnlyToCustomer

		if b.BGPPathA.NextHop != nil {
			a.NextHop = b.BGPPathA.NextHop.ToProto()
		}

		if b.BGPPathA.Source != nil {
			a.Source = b.BGPPathA.Source.ToProto()
		}
	}

	if b.ASPath != nil {
		a.AsPath = b.ASPath.ToProto()
	}

	if a.ClusterList != nil {
		a.ClusterList = make([]uint32, len(*b.ClusterList))
		for i := range *b.ClusterList {
			a.ClusterList[i] = (*b.ClusterList)[i]
		}
	}

	if b.Communities != nil {
		a.Communities = make([]uint32, len(*b.Communities))
		for i := range *b.Communities {
			a.Communities[i] = (*b.Communities)[i]
		}
	}

	if b.LargeCommunities != nil {
		a.LargeCommunities = make([]*api.LargeCommunity, len(*b.LargeCommunities))
		for i := range *b.LargeCommunities {
			a.LargeCommunities[i] = (*b.LargeCommunities)[i].ToProto()
		}
	}

	for i := range b.UnknownAttributes {
		a.UnknownAttributes[i] = b.UnknownAttributes[i].ToProto()
	}

	return a
}

// BGPPathFromProtoBGPPath converts a proto BGPPath to BGPPath
func BGPPathFromProtoBGPPath(pb *api.BGPPath, dedup bool) *BGPPath {
	p := &BGPPath{
		BGPPathA: &BGPPathA{
			NextHop:        bnet.IPFromProtoIP(pb.NextHop).Ptr(),
			LocalPref:      pb.LocalPref,
			OriginatorID:   pb.OriginatorId,
			Origin:         uint8(pb.Origin),
			MED:            pb.Med,
			EBGP:           pb.Ebgp,
			BGPIdentifier:  pb.BgpIdentifier,
			Source:         bnet.IPFromProtoIP(pb.Source).Ptr(),
			OnlyToCustomer: pb.OnlyToCustomer,
		},
		PathIdentifier: pb.PathIdentifier,
		ASPath:         types.ASPathFromProtoASPath(pb.AsPath),
		BMPPostPolicy:  pb.BmpPostPolicy,
	}

	if dedup {
		p = p.Dedup()
	}

	communities := make(types.Communities, len(pb.Communities))
	p.Communities = &communities

	largeCommunities := make(types.LargeCommunities, len(pb.LargeCommunities))
	p.LargeCommunities = &largeCommunities

	unknownAttr := make([]types.UnknownPathAttribute, len(pb.UnknownAttributes))
	p.UnknownAttributes = unknownAttr

	cl := make(types.ClusterList, len(pb.ClusterList))
	p.ClusterList = &cl

	for i := range pb.Communities {
		(*p.Communities)[i] = pb.Communities[i]
	}

	for i := range pb.LargeCommunities {
		(*p.LargeCommunities)[i] = types.LargeCommunityFromProtoCommunity(pb.LargeCommunities[i])
	}

	for i := range pb.UnknownAttributes {
		p.UnknownAttributes[i] = types.UnknownPathAttributeFromProtoUnknownPathAttribute(pb.UnknownAttributes[i])
	}

	for i := range pb.ClusterList {
		(*p.ClusterList)[i] = pb.ClusterList[i]
	}

	return p
}

// Length get's the length of serialized path
func (b *BGPPath) Length() uint16 {
	asPathLen := uint16(3)
	for _, segment := range *b.ASPath {
		asPathLen++
		asPathLen += uint16(4 * len(segment.ASNs))
	}

	communitiesLen := uint16(0)
	if b.Communities != nil && len(*b.Communities) != 0 {
		communitiesLen += 3 + uint16(len(*b.Communities)*4)
	}

	largeCommunitiesLen := uint16(0)
	if b.LargeCommunities != nil && len(*b.LargeCommunities) != 0 {
		largeCommunitiesLen += 3 + uint16(len(*b.LargeCommunities)*12)
	}

	clusterListLen := uint16(0)
	if b.ClusterList != nil && len(*b.ClusterList) != 0 {
		clusterListLen += 3 + uint16(len(*b.ClusterList)*4)
	}

	originatorID := uint16(0)
	if b.BGPPathA.OriginatorID != 0 {
		originatorID = 4
	}

	onlyToCustomer := uint16(0)
	if b.BGPPathA.OnlyToCustomer != 0 {
		onlyToCustomer = 4
	}

	unknownAttributesLen := uint16(0)
	if b.UnknownAttributes != nil {
		for _, unknownAttr := range b.UnknownAttributes {
			unknownAttributesLen += unknownAttr.WireLength()
		}
	}

	return 4*7 + 4 + asPathLen + communitiesLen + largeCommunitiesLen + clusterListLen + originatorID + onlyToCustomer + unknownAttributesLen
}

// ECMP determines if routes b and c are euqal in terms of ECMP
func (b *BGPPath) ECMP(c *BGPPath) bool {
	return b.BGPPathA.LocalPref == c.BGPPathA.LocalPref &&
		b.ASPathLen == c.ASPathLen &&
		b.BGPPathA.MED == c.BGPPathA.MED &&
		b.BGPPathA.Origin == c.BGPPathA.Origin
}

// Compare checks if paths are the same
func (b *BGPPath) Compare(c *BGPPath) bool {
	if b.PathIdentifier != c.PathIdentifier {
		return false
	}

	if !b.BGPPathA.compare(c.BGPPathA) {
		return false
	}

	if !b.ASPath.Compare(c.ASPath) {
		return false
	}

	if !b.compareClusterList(c) {
		return false
	}

	if !b.compareCommunities(c) {
		return false
	}

	if !b.compareLargeCommunities(c) {
		return false
	}

	if !b.compareUnknownAttributes(c) {
		return false
	}

	return true
}

func (b *BGPPath) compareCommunities(c *BGPPath) bool {
	if b.Communities == nil && c.Communities == nil {
		return true
	}

	if b.Communities != nil && c.Communities == nil {
		return false
	}

	if b.Communities == nil && c.Communities != nil {
		return false
	}

	if len(*b.Communities) != len(*c.Communities) {
		return false
	}

	for i := range *b.Communities {
		if (*b.Communities)[i] != (*c.Communities)[i] {
			return false
		}
	}

	return true
}

func (b *BGPPath) compareClusterList(c *BGPPath) bool {
	if b.ClusterList == nil && c.ClusterList == nil {
		return true
	}

	if b.ClusterList != nil && c.ClusterList == nil {
		return false
	}

	if b.ClusterList == nil && c.ClusterList != nil {
		return false
	}

	if len(*b.ClusterList) != len(*c.ClusterList) {
		return false
	}

	for i := range *b.ClusterList {
		if (*b.ClusterList)[i] != (*c.ClusterList)[i] {
			return false
		}
	}

	return true
}

func (b *BGPPath) compareLargeCommunities(c *BGPPath) bool {
	if b.LargeCommunities == nil && c.LargeCommunities == nil {
		return true
	}

	if b.LargeCommunities != nil && c.LargeCommunities == nil {
		return false
	}

	if b.LargeCommunities == nil && c.LargeCommunities != nil {
		return false
	}

	if len(*b.LargeCommunities) != len(*c.LargeCommunities) {
		return false
	}

	for i := range *b.LargeCommunities {
		if (*b.LargeCommunities)[i] != (*c.LargeCommunities)[i] {
			return false
		}
	}

	return true
}

func (b *BGPPath) compareUnknownAttributes(c *BGPPath) bool {
	if len(b.UnknownAttributes) != len(c.UnknownAttributes) {
		return false
	}

	for i := range b.UnknownAttributes {
		if !b.UnknownAttributes[i].Compare(&c.UnknownAttributes[i]) {
			return false
		}
	}

	return true
}

func (b *BGPPathA) compare(c *BGPPathA) bool {
	if b.NextHop.Compare(c.NextHop) != 0 {
		return false
	}

	if b.Source.Compare(c.Source) != 0 {
		return false
	}

	if b.LocalPref != c.LocalPref || b.MED != c.MED || b.BGPIdentifier != c.BGPIdentifier || b.OriginatorID != c.OriginatorID {
		return false
	}

	if b.EBGP != c.EBGP || b.AtomicAggregate != c.AtomicAggregate || b.Origin != c.Origin {
		return false
	}

	if b.Aggregator != nil || c.Aggregator != nil {
		if b.Aggregator != nil && c.Aggregator != nil {
			if *b.Aggregator != *c.Aggregator {
				return false
			}
		} else {
			return false
		}
	}

	return true
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
	if c.BGPPathA.LocalPref < b.BGPPathA.LocalPref {
		return 1
	}

	if c.BGPPathA.LocalPref > b.BGPPathA.LocalPref {
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
	if c.BGPPathA.Origin > b.BGPPathA.Origin {
		return 1
	}

	if c.BGPPathA.Origin < b.BGPPathA.Origin {
		return -1
	}

	// c)
	if c.BGPPathA.MED > b.BGPPathA.MED {
		return 1
	}

	if c.BGPPathA.MED < b.BGPPathA.MED {
		return -1
	}

	// d)
	if c.BGPPathA.EBGP && !b.BGPPathA.EBGP {
		return -1
	}

	if !c.BGPPathA.EBGP && b.BGPPathA.EBGP {
		return 1
	}

	// e) TODO: interior cost (hello IS-IS and OSPF)

	// f) + RFC4456 9. (Route Reflection)
	bgpIdentifierC := c.BGPPathA.BGPIdentifier
	bgpIdentifierB := b.BGPPathA.BGPIdentifier

	// IF an OriginatorID (set by an RR) is present, use this instead of Originator
	if c.BGPPathA.OriginatorID != 0 {
		bgpIdentifierC = c.BGPPathA.OriginatorID
	}

	if b.BGPPathA.OriginatorID != 0 {
		bgpIdentifierB = b.BGPPathA.OriginatorID
	}

	if bgpIdentifierC < bgpIdentifierB {
		return 1
	}

	if bgpIdentifierC > bgpIdentifierB {
		return -1
	}

	if c.ClusterList != nil && b.ClusterList != nil {
		// Additionally check for the shorter ClusterList
		if len(*c.ClusterList) < len(*b.ClusterList) {
			return 1
		}

		if len(*c.ClusterList) > len(*b.ClusterList) {
			return -1
		}
	}

	// g)
	if c.BGPPathA.Source.Compare(b.BGPPathA.Source) == -1 {
		return 1
	}

	if c.BGPPathA.Source.Compare(b.BGPPathA.Source) == 1 {
		return -1
	}

	if c.BGPPathA.NextHop.Compare(b.BGPPathA.NextHop) == -1 {
		return 1
	}

	if c.BGPPathA.NextHop.Compare(b.BGPPathA.NextHop) == 1 {
		return -1
	}

	return 0
}

func (b *BGPPath) betterECMP(c *BGPPath) bool {
	if c.BGPPathA.LocalPref < b.BGPPathA.LocalPref {
		return false
	}

	if c.BGPPathA.LocalPref > b.BGPPathA.LocalPref {
		return true
	}

	if c.ASPathLen > b.ASPathLen {
		return false
	}

	if c.ASPathLen < b.ASPathLen {
		return true
	}

	if c.BGPPathA.Origin > b.BGPPathA.Origin {
		return false
	}

	if c.BGPPathA.Origin < b.BGPPathA.Origin {
		return true
	}

	if c.BGPPathA.MED > b.BGPPathA.MED {
		return false
	}

	if c.BGPPathA.MED < b.BGPPathA.MED {
		return true
	}

	return false
}

func (b *BGPPath) better(c *BGPPath) bool {
	if b.betterECMP(c) {
		return true
	}

	if c.BGPPathA.BGPIdentifier < b.BGPPathA.BGPIdentifier {
		return true
	}

	if c.BGPPathA.Source.Compare(b.BGPPathA.Source) == -1 {
		return true
	}

	return false
}

// Print all known information about a route in logfile friendly format
func (b *BGPPath) String() string {
	buf := &strings.Builder{}

	origin := ""
	switch b.BGPPathA.Origin {
	case 0:
		origin = "IGP"
	case 1:
		origin = "EGP"
	case 2:
		origin = "Incomplete"
	}

	bgpType := "internal"
	if b.BGPPathA.EBGP {
		bgpType = "external"
	}

	fmt.Fprintf(buf, "Local Pref: %d, ", b.BGPPathA.LocalPref)
	fmt.Fprintf(buf, "Origin: %s, ", origin)
	fmt.Fprintf(buf, "AS Path: %v, ", b.ASPath)
	fmt.Fprintf(buf, "BGP type: %s, ", bgpType)
	fmt.Fprintf(buf, "NEXT HOP: %s, ", b.BGPPathA.NextHop)
	fmt.Fprintf(buf, "MED: %d, ", b.BGPPathA.MED)
	fmt.Fprintf(buf, "Path ID: %d, ", b.PathIdentifier)
	fmt.Fprintf(buf, "Source: %s, ", b.BGPPathA.Source)
	if b.BGPPathA.OnlyToCustomer != 0 {
		fmt.Fprintf(buf, "OnlyToCustomer: %d, ", b.BGPPathA.OnlyToCustomer)
	}
	if b.Communities != nil {
		fmt.Fprintf(buf, "Communities: %v, ", *b.Communities)
	}
	if b.LargeCommunities != nil {
		fmt.Fprintf(buf, "LargeCommunities: %v", *b.LargeCommunities)
	}

	if b.BGPPathA.OriginatorID != 0 {
		oid := convert.Uint32Byte(b.BGPPathA.OriginatorID)
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
	switch b.BGPPathA.Origin {
	case 0:
		origin = "IGP"
	case 1:
		origin = "EGP"
	case 2:
		origin = "Incomplete"
	}

	bgpType := "internal"
	if b.BGPPathA.EBGP {
		bgpType = "external"
	}

	fmt.Fprintf(buf, "\t\tLocal Pref: %d\n", b.BGPPathA.LocalPref)
	fmt.Fprintf(buf, "\t\tOrigin: %s\n", origin)
	fmt.Fprintf(buf, "\t\tAS Path: %v\n", b.ASPath)
	fmt.Fprintf(buf, "\t\tBGP type: %s\n", bgpType)
	fmt.Fprintf(buf, "\t\tNEXT HOP: %s\n", b.BGPPathA.NextHop)
	fmt.Fprintf(buf, "\t\tMED: %d\n", b.BGPPathA.MED)
	fmt.Fprintf(buf, "\t\tPath ID: %d\n", b.PathIdentifier)
	fmt.Fprintf(buf, "\t\tSource: %s\n", b.BGPPathA.Source)
	if b.BGPPathA.OnlyToCustomer != 0 {
		fmt.Fprintf(buf, "\t\tOnlyToCustomer: %d\n", b.BGPPathA.OnlyToCustomer)
	}
	if b.Communities != nil {
		fmt.Fprintf(buf, "\t\tCommunities: %v\n", *b.Communities)
	}
	if b.LargeCommunities != nil {
		fmt.Fprintf(buf, "\t\tLargeCommunities: %v\n", *b.LargeCommunities)
	}

	if b.BGPPathA.OriginatorID != 0 {
		oid := convert.Uint32Byte(b.BGPPathA.OriginatorID)
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

	if len(*b.ASPath) == 0 {
		b.insertNewASSequence()
	}

	first := (*b.ASPath)[0]
	if first.Type == types.ASSet {
		b.insertNewASSequence()
	}

	for i := 0; i < int(times); i++ {
		if len(*b.ASPath) == types.MaxASNsSegment {
			b.insertNewASSequence()
		}

		old := (*b.ASPath)[0].ASNs
		asns := make([]uint32, len(old)+1)
		copy(asns[1:], old)
		asns[0] = asn
		(*b.ASPath)[0].ASNs = asns
	}

	b.ASPathLen = b.ASPath.Length()
}

func (b *BGPPath) insertNewASSequence() {
	pa := make(types.ASPath, len(*b.ASPath)+1)
	copy(pa[1:], (*b.ASPath))
	pa[0] = types.ASPathSegment{
		ASNs: make([]uint32, 0),
		Type: types.ASSequence,
	}

	b.ASPath = &pa
}

// Copy creates a deep copy of a BGPPath
func (b *BGPPath) Copy() *BGPPath {
	if b == nil {
		return nil
	}

	cp := *b

	if cp.ASPath != nil {
		asPath := make(types.ASPath, len(*cp.ASPath))
		cp.ASPath = &asPath
		copy(*cp.ASPath, *b.ASPath)
	}

	if cp.Communities != nil {
		communities := make(types.Communities, len(*cp.Communities))
		cp.Communities = &communities
		copy(*cp.Communities, *b.Communities)
	}

	if cp.LargeCommunities != nil {
		largeCommunities := make(types.LargeCommunities, len(*cp.LargeCommunities))
		cp.LargeCommunities = &largeCommunities
		copy(*cp.LargeCommunities, *b.LargeCommunities)
	}

	if b.ClusterList != nil {
		clusterList := make(types.ClusterList, len(*cp.ClusterList))
		cp.ClusterList = &clusterList
		copy(*cp.ClusterList, *b.ClusterList)
	}

	return &cp
}

// ComputeHash computes an hash over all attributes of the path
func (b *BGPPath) ComputeHash() string {
	s := fmt.Sprintf("%s\t%d\t%s\t%d\t%d\t%v\t%d\t%s\t%s\t%s\t%d\t%s",
		b.BGPPathA.NextHop.String(),
		b.BGPPathA.LocalPref,
		b.ASPath.String(),
		b.BGPPathA.Origin,
		b.BGPPathA.MED,
		b.BGPPathA.EBGP,
		b.BGPPathA.BGPIdentifier,
		b.BGPPathA.Source.String(),
		b.Communities.String(),
		b.LargeCommunities.String(),
		b.BGPPathA.OriginatorID,
		b.ClusterList.String())

	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

// ComputeHash computes an hash over all attributes of the path
func (b *BGPPath) ComputeHashWithPathID() string {
	s := fmt.Sprintf("%s\t%d\t%s\t%d\t%d\t%v\t%d\t%s\t%s\t%s\t%d\t%d\t%s",
		b.BGPPathA.NextHop.String(),
		b.BGPPathA.LocalPref,
		b.ASPath.String(),
		b.BGPPathA.Origin,
		b.BGPPathA.MED,
		b.BGPPathA.EBGP,
		b.BGPPathA.BGPIdentifier,
		b.BGPPathA.Source.String(),
		b.Communities.String(),
		b.LargeCommunities.String(),
		b.PathIdentifier,
		b.BGPPathA.OriginatorID,
		b.ClusterList.String())

	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

// CommunitiesString returns the formated communities
func (b *BGPPath) CommunitiesString() string {
	str := &strings.Builder{}

	for i, com := range *b.Communities {
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

	for i, cid := range *b.ClusterList {
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

	for i, com := range *b.LargeCommunities {
		if i > 0 {
			str.WriteByte(' ')
		}
		str.WriteString(com.String())
	}

	return str.String()
}
