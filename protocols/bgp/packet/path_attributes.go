package packet

import (
	"bytes"
	"fmt"
	"math"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"
)

func decodePathAttrs(buf *bytes.Buffer, tpal uint16, opt *DecodeOptions) (*PathAttribute, error) {
	if tpal == 0 {
		return nil, nil
	}

	var ret *PathAttribute
	var eol *PathAttribute
	var pa *PathAttribute
	var err error
	var consumed uint16

	haveNextHop := false
	haveOrigin := false
	haveASPath := false

	p := uint16(0)
	for p < tpal {
		pa, consumed, err = decodePathAttr(buf, opt)
		if err != nil {
			return nil, fmt.Errorf("unable to decode path attr: %w", err)
		}
		p += consumed

		switch pa.TypeCode {
		case ASPathAttr:
			haveASPath = true
		case OriginAttr:
			haveOrigin = true
		case NextHopAttr:
			haveNextHop = true
		case MultiProtocolReachNLRIAttr:
			haveNextHop = true
		}

		if ret == nil {
			ret = pa
			eol = pa
		} else {
			eol.Next = pa
			eol = pa
		}
	}

	if haveNextHop || haveOrigin || haveASPath {
		if !haveNextHop {
			return nil, fmt.Errorf("next-hop attribute missing")
		}

		if !haveOrigin {
			return nil, fmt.Errorf("origin attribute missing")
		}

		if !haveASPath {
			return nil, fmt.Errorf("AS path attribute missing")
		}
	}

	return ret, nil
}

func decodePathAttr(buf *bytes.Buffer, opt *DecodeOptions) (pa *PathAttribute, consumed uint16, err error) {
	pa = &PathAttribute{}

	err = decodePathAttrFlags(buf, pa)
	if err != nil {
		return nil, consumed, fmt.Errorf("unable to get path attribute flags: %w", err)
	}
	consumed++

	err = decode.DecodeUint8(buf, &pa.TypeCode)
	if err != nil {
		return nil, consumed, err
	}
	consumed++

	n, err := pa.setLength(buf)
	if err != nil {
		return nil, consumed, err
	}
	consumed += uint16(n)

	switch pa.TypeCode {
	case OriginAttr:
		if err := pa.decodeOrigin(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode Origin: %w", err)
		}
	case ASPathAttr:
		asnLength := uint8(2)
		if opt.Use32BitASN {
			asnLength = 4
		}

		if err := pa.decodeASPath(buf, asnLength); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode AS Path: %w", err)
		}
	/* Don't decodeAS4Paths yet: The rest of the software does not support it right yet!
	case AS4PathAttr:
		if err := pa.decodeASPath(buf, 4); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode AS4 Path: %w", err)
		}*/
	case NextHopAttr:
		if err := pa.decodeNextHop(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode Next-Hop: %w", err)
		}
	case MEDAttr:
		if err := pa.decodeMED(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode MED: %w", err)
		}
	case LocalPrefAttr:
		if err := pa.decodeLocalPref(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode local pref: %w", err)
		}
	case AggregatorAttr:
		if err := pa.decodeAggregator(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode Aggregator: %w", err)
		}
	case AtomicAggrAttr:
		// Nothing to do for 0 octet long attribute
	case CommunitiesAttr:
		if err := pa.decodeCommunities(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode Community: %w", err)
		}
	case OriginatorIDAttr:
		if err := pa.decodeOriginatorID(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode OriginatorID: %w", err)
		}
	case ClusterListAttr:
		if err := pa.decodeClusterList(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode OriginatorID: %w", err)
		}
	case MultiProtocolReachNLRIAttr:
		if err := pa.decodeMultiProtocolReachNLRI(buf, opt); err != nil {
			return nil, consumed, fmt.Errorf("Failed to multi protocol reachable NLRI: %w", err)
		}
	case MultiProtocolUnreachNLRIAttr:
		if err := pa.decodeMultiProtocolUnreachNLRI(buf, opt); err != nil {
			return nil, consumed, fmt.Errorf("Failed to multi protocol unreachable NLRI: %w", err)
		}
	case AS4AggregatorAttr:
		if err := pa.decodeAS4Aggregator(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to skip not supported AS4Aggregator: %w", err)
		}
	case LargeCommunitiesAttr:
		if err := pa.decodeLargeCommunities(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode large communities: %w", err)
		}
	default:
		if err := pa.decodeUnknown(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode unknown attribute: %w", err)
		}
	}

	return pa, consumed + pa.Length, nil
}

func (pa *PathAttribute) decodeMultiProtocolReachNLRI(buf *bytes.Buffer, opt *DecodeOptions) error {
	b := make([]byte, pa.Length)
	n, err := buf.Read(b)
	if err != nil {
		return fmt.Errorf("unable to read %d bytes from buffer: %w", pa.Length, err)
	}
	if n != int(pa.Length) {
		return fmt.Errorf("unable to read %d bytes from buffer, only got %d bytes", pa.Length, n)
	}

	nlri, err := deserializeMultiProtocolReachNLRI(b, opt)
	if err != nil {
		return fmt.Errorf("unable to decode MP_REACH_NLRI: %w", err)
	}

	pa.Value = nlri
	return nil
}

func (pa *PathAttribute) decodeMultiProtocolUnreachNLRI(buf *bytes.Buffer, opt *DecodeOptions) error {
	b := make([]byte, pa.Length)
	n, err := buf.Read(b)
	if err != nil {
		return fmt.Errorf("unable to read %d bytes from buffer: %w", pa.Length, err)
	}
	if n != int(pa.Length) {
		return fmt.Errorf("unable to read %d bytes from buffer, only got %d bytes", pa.Length, n)
	}

	nlri, err := deserializeMultiProtocolUnreachNLRI(b, opt)
	if err != nil {
		return fmt.Errorf("unable to decode MP_UNREACH_NLRI: %w", err)
	}

	pa.Value = nlri
	return nil
}

func (pa *PathAttribute) decodeUnknown(buf *bytes.Buffer) error {
	u := make([]byte, pa.Length)

	err := decode.Decode(buf, []interface{}{&u})
	if err != nil {
		return fmt.Errorf("unable to decode: %w", err)
	}

	pa.Value = u
	return nil
}

func (pa *PathAttribute) decodeOrigin(buf *bytes.Buffer) error {
	origin := uint8(0)

	p := uint16(0)
	err := decode.DecodeUint8(buf, &origin)
	if err != nil {
		return fmt.Errorf("unable to decode: %w", err)
	}

	pa.Value = origin
	p++

	return dumpNBytes(buf, pa.Length-p)
}

func (pa *PathAttribute) decodeASPath(buf *bytes.Buffer, asnLength uint8) error {
	pa.Value = make(types.ASPath, 0, 1)
	p := uint16(0)
	for p < pa.Length {
		segment := types.ASPathSegment{}
		count := uint8(0)

		err := decode.DecodeUint8(buf, &segment.Type)
		if err != nil {
			return err
		}

		err = decode.DecodeUint8(buf, &count)
		if err != nil {
			return err
		}

		p += 2

		if segment.Type != types.ASSet && segment.Type != types.ASSequence {
			return fmt.Errorf("Invalid AS Path segment type: %d", segment.Type)
		}

		if count == 0 {
			return fmt.Errorf("Invalid AS Path segment length: %d", count)
		}

		segment.ASNs = make([]uint32, count)
		for i := uint8(0); i < count; i++ {
			asn, err := pa.decodeASN(buf, asnLength)
			if err != nil {
				return err
			}
			p += uint16(asnLength)

			segment.ASNs[i] = asn
		}

		pa.Value = append(pa.Value.(types.ASPath), segment)
	}

	ret := pa.Value.(types.ASPath)
	pa.Value = &ret
	return nil
}

func (pa *PathAttribute) decodeASN(buf *bytes.Buffer, asnSize uint8) (asn uint32, err error) {
	if asnSize == 4 {
		return decode4ByteASN(buf)
	}

	return decode2ByteASN(buf)
}

func decode4ByteASN(buf *bytes.Buffer) (asn uint32, err error) {
	asn4 := uint32(0)
	err = decode.DecodeUint32(buf, &asn4)
	if err != nil {
		return 0, err
	}

	return asn4, nil
}

func decode2ByteASN(buf *bytes.Buffer) (asn uint32, err error) {
	asn2 := uint16(0)
	err = decode.DecodeUint16(buf, &asn2)
	if err != nil {
		return 0, err
	}

	return uint32(asn2), nil
}

func (pa *PathAttribute) decodeNextHop(buf *bytes.Buffer) error {
	nextHop := uint32(0)
	err := decode.Decode(buf, []interface{}{&nextHop})
	if err != nil {
		return fmt.Errorf("unable to decode next hop: %w", err)
	}

	pa.Value = bnet.IPv4(nextHop).Dedup()
	return nil
}

func (pa *PathAttribute) decodeMED(buf *bytes.Buffer) error {
	med := uint32(0)
	err := decode.DecodeUint32(buf, &med)
	if err != nil {
		return err
	}

	pa.Value = med
	return nil
}

func (pa *PathAttribute) decodeLocalPref(buf *bytes.Buffer) error {
	lp := uint32(0)
	err := decode.DecodeUint32(buf, &lp)
	if err != nil {
		return err
	}

	pa.Value = lp
	return nil
}

func (pa *PathAttribute) decodeAggregator(buf *bytes.Buffer) error {
	aggr := types.Aggregator{}
	p := uint16(0)

	err := decode.Decode(buf, []interface{}{&aggr.ASN, &aggr.Address})
	if err != nil {
		return err
	}
	p += 6
	pa.Value = aggr
	return dumpNBytes(buf, pa.Length-p)
}

func (pa *PathAttribute) decodeCommunities(buf *bytes.Buffer) error {
	if pa.Length%CommunityLen != 0 {
		return fmt.Errorf("unable to read community path attribute. Length %d is not divisible by 4", pa.Length)
	}

	count := pa.Length / CommunityLen
	coms := make(types.Communities, count)

	for i := uint16(0); i < count; i++ {
		v, err := read4BytesAsUint32(buf)
		if err != nil {
			return err
		}
		coms[i] = v
	}

	pa.Value = &coms
	return nil
}

func (pa *PathAttribute) decodeLargeCommunities(buf *bytes.Buffer) error {
	if pa.Length%LargeCommunityLen != 0 {
		return fmt.Errorf("unable to read large community path attribute. Length %d is not divisible by 12", pa.Length)
	}

	count := pa.Length / LargeCommunityLen
	coms := make(types.LargeCommunities, count)

	for i := uint16(0); i < count; i++ {
		com := types.LargeCommunity{}

		v, err := read4BytesAsUint32(buf)
		if err != nil {
			return err
		}
		com.GlobalAdministrator = v

		v, err = read4BytesAsUint32(buf)
		if err != nil {
			return err
		}
		com.DataPart1 = v

		v, err = read4BytesAsUint32(buf)
		if err != nil {
			return err
		}
		com.DataPart2 = v

		coms[i] = com
	}

	pa.Value = &coms
	return nil
}

func (pa *PathAttribute) decodeAS4Aggregator(buf *bytes.Buffer) error {
	return pa.decodeUint32(buf, "AS4Aggregator")
}

func (pa *PathAttribute) decodeUint32(buf *bytes.Buffer, attrName string) error {
	v, err := read4BytesAsUint32(buf)
	if err != nil {
		return fmt.Errorf("unable to decode %s: %w", attrName, err)
	}

	pa.Value = v

	p := uint16(4)
	err = dumpNBytes(buf, pa.Length-p)
	if err != nil {
		return fmt.Errorf("dumpNBytes failed: %w", err)
	}

	return nil
}

func (pa *PathAttribute) decodeOriginatorID(buf *bytes.Buffer) error {
	return pa.decodeUint32(buf, "OriginatorID")
}

func (pa *PathAttribute) decodeClusterList(buf *bytes.Buffer) error {
	if pa.Length%ClusterIDLen != 0 {
		return fmt.Errorf("unable to read ClusterList path attribute. Length %d is not divisible by %d", pa.Length, ClusterIDLen)
	}

	count := pa.Length / ClusterIDLen
	cids := make(types.ClusterList, count)

	for i := uint16(0); i < count; i++ {
		v, err := read4BytesAsUint32(buf)
		if err != nil {
			return err
		}
		cids[i] = v
	}

	pa.Value = &cids
	return nil
}

func (pa *PathAttribute) setLength(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	if pa.ExtendedLength {
		err := decode.DecodeUint16(buf, &pa.Length)
		if err != nil {
			return 0, err
		}

		bytesRead = 2
	} else {
		x := uint8(0)
		err := decode.DecodeUint8(buf, &x)
		if err != nil {
			return 0, err
		}

		pa.Length = uint16(x)
		bytesRead = 1
	}
	return bytesRead, nil
}

// Copy create a copy of a path attribute
func (pa *PathAttribute) Copy() *PathAttribute {
	return &PathAttribute{
		ExtendedLength: pa.ExtendedLength,
		Length:         pa.Length,
		Optional:       pa.Optional,
		Partial:        pa.Partial,
		Transitive:     pa.Transitive,
		TypeCode:       pa.TypeCode,
		Value:          pa.Value,
	}
}

// dumpNBytes is used to dump n bytes of buf. This is useful in case an path attributes
// length doesn't match a fixed length's attributes length (e.g. ORIGIN is always an octet)
func dumpNBytes(buf *bytes.Buffer, n uint16) error {
	if n <= 0 {
		return nil
	}

	for i := uint16(0); i < n; i++ {
		_, err := buf.ReadByte()
		if err != nil {
			return err
		}
	}

	return nil
}

// Serialize serializes a path attribute
func (pa *PathAttribute) Serialize(buf *bytes.Buffer, opt *EncodeOptions) uint16 {
	pathAttrLen := uint16(0)

	switch pa.TypeCode {
	case OriginAttr:
		pathAttrLen = uint16(pa.serializeOrigin(buf))
	case ASPathAttr:
		pathAttrLen = pa.serializeASPath(buf, opt)
	case NextHopAttr:
		pathAttrLen = uint16(pa.serializeNextHop(buf))
	case MEDAttr:
		pathAttrLen = uint16(pa.serializeMED(buf))
	case LocalPrefAttr:
		pathAttrLen = uint16(pa.serializeLocalpref(buf))
	case AtomicAggrAttr:
		pathAttrLen = uint16(pa.serializeAtomicAggregate(buf))
	case AggregatorAttr:
		pathAttrLen = uint16(pa.serializeAggregator(buf))
	case CommunitiesAttr:
		pathAttrLen = uint16(pa.serializeCommunities(buf))
	case LargeCommunitiesAttr:
		pathAttrLen = uint16(pa.serializeLargeCommunities(buf))
	case MultiProtocolReachNLRIAttr:
		pathAttrLen = pa.serializeMultiProtocolReachNLRI(buf, opt)
	case MultiProtocolUnreachNLRIAttr:
		pathAttrLen = pa.serializeMultiProtocolUnreachNLRI(buf, opt)
	case OriginatorIDAttr:
		pathAttrLen = uint16(pa.serializeOriginatorID(buf))
	case ClusterListAttr:
		pathAttrLen = uint16(pa.serializeClusterList(buf))
	default:
		pathAttrLen = pa.serializeUnknownAttribute(buf)
	}

	return pathAttrLen
}

func (pa *PathAttribute) serializeOrigin(buf *bytes.Buffer) uint8 {
	attrFlags := uint8(0)
	attrFlags = setTransitive(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(OriginAttr)
	length := uint8(1)
	buf.WriteByte(length)
	buf.WriteByte(pa.Value.(uint8))
	return length + 3
}

func (pa *PathAttribute) serializeASPath(buf *bytes.Buffer, opt *EncodeOptions) uint16 {
	attrFlags := uint8(0)
	attrFlags = setTransitive(attrFlags)

	asnLength := uint16(2)
	if opt.Use32BitASN {
		asnLength = 4
	}

	length := uint16(0)
	segmentsBuf := bytes.NewBuffer(nil)
	for _, segment := range *pa.Value.(*types.ASPath) {
		segmentsBuf.WriteByte(segment.Type)
		segmentsBuf.WriteByte(uint8(len(segment.ASNs)))

		for _, asn := range segment.ASNs {
			if opt.Use32BitASN {
				segmentsBuf.Write(convert.Uint32Byte(asn))
			} else {
				segmentsBuf.Write(convert.Uint16Byte(uint16(asn)))
			}
		}
		length += 2 + uint16(len(segment.ASNs))*asnLength
	}

	if length > 255 {
		attrFlags = setExtendedLength(attrFlags)
	}

	buf.WriteByte(attrFlags)
	buf.WriteByte(ASPathAttr)
	if length < 256 {
		buf.WriteByte(uint8(length))
	} else {
		buf.Write(convert.Uint16Byte(length))
	}

	buf.Write(segmentsBuf.Bytes())

	return length + 3
}

func (pa *PathAttribute) serializeNextHop(buf *bytes.Buffer) uint8 {
	attrFlags := uint8(0)
	attrFlags = setTransitive(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(NextHopAttr)
	length := uint8(4)
	buf.WriteByte(length)
	addr := pa.Value.(*bnet.IP)
	buf.Write(addr.Bytes())
	return length + 3
}

func (pa *PathAttribute) serializeMED(buf *bytes.Buffer) uint8 {
	attrFlags := uint8(0)
	attrFlags = setOptional(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(MEDAttr)
	length := uint8(4)
	buf.WriteByte(length)
	buf.Write(convert.Uint32Byte(pa.Value.(uint32)))
	return length + 3
}

func (pa *PathAttribute) serializeLocalpref(buf *bytes.Buffer) uint8 {
	attrFlags := uint8(0)
	attrFlags = setTransitive(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(LocalPrefAttr)
	length := uint8(4)
	buf.WriteByte(length)
	buf.Write(convert.Uint32Byte(pa.Value.(uint32)))
	return length + 3
}

func (pa *PathAttribute) serializeAtomicAggregate(buf *bytes.Buffer) uint8 {
	attrFlags := uint8(0)
	attrFlags = setTransitive(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(AtomicAggrAttr)
	length := uint8(0)
	buf.WriteByte(length)
	return 3
}

func (pa *PathAttribute) serializeAggregator(buf *bytes.Buffer) uint8 {
	attrFlags := uint8(0)
	attrFlags = setOptional(attrFlags)
	attrFlags = setTransitive(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(AggregatorAttr)
	length := uint8(6)
	buf.WriteByte(length)

	aggregator := pa.Value.(types.Aggregator)
	buf.Write(convert.Uint16Byte(aggregator.ASN))
	buf.Write(convert.Uint32Byte(aggregator.Address))

	return length + 3
}

func (pa *PathAttribute) serializeCommunities(buf *bytes.Buffer) uint16 {
	if pa.Value == nil {
		return 0
	}

	coms := pa.Value.(*types.Communities)
	if len(*coms) == 0 {
		return 0
	}

	length := uint16(CommunityLen * len(*coms))

	attrFlags := uint8(0)
	attrFlags = setOptional(attrFlags)
	attrFlags = setTransitive(attrFlags)
	attrFlags = setPartial(attrFlags)
	if length > 255 {
		attrFlags = setExtendedLength(attrFlags)
	}
	buf.WriteByte(attrFlags)
	buf.WriteByte(CommunitiesAttr)

	if length < 256 {
		buf.WriteByte(uint8(length))
	} else {
		buf.Write(convert.Uint16Byte(length))
		length++
	}

	for _, com := range *coms {
		buf.Write(convert.Uint32Byte(com))
	}

	return length + 3
}

func (pa *PathAttribute) serializeLargeCommunities(buf *bytes.Buffer) uint16 {
	if pa.Value == nil {
		return 0
	}

	coms := pa.Value.(*types.LargeCommunities)
	if len(*coms) == 0 {
		return 0
	}

	length := uint16(LargeCommunityLen * len(*coms))

	attrFlags := uint8(0)
	attrFlags = setOptional(attrFlags)
	attrFlags = setTransitive(attrFlags)
	attrFlags = setPartial(attrFlags)
	if length > 255 {
		attrFlags = setExtendedLength(attrFlags)
	}
	buf.WriteByte(attrFlags)
	buf.WriteByte(LargeCommunitiesAttr)

	if length < 256 {
		buf.WriteByte(uint8(length))
	} else {
		buf.Write(convert.Uint16Byte(length))
		length++
	}

	for _, com := range *coms {
		buf.Write(convert.Uint32Byte(com.GlobalAdministrator))
		buf.Write(convert.Uint32Byte(com.DataPart1))
		buf.Write(convert.Uint32Byte(com.DataPart2))
	}

	return length + 3
}

func (pa *PathAttribute) serializeOriginatorID(buf *bytes.Buffer) uint8 {
	attrFlags := uint8(0)
	attrFlags = setOptional(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(OriginatorIDAttr)
	length := uint8(4)
	buf.WriteByte(length)
	oid := pa.Value.(uint32)
	buf.Write(convert.Uint32Byte(oid))
	return 7
}

func (pa *PathAttribute) serializeClusterList(buf *bytes.Buffer) uint8 {
	if pa.Value == nil {
		return 0
	}

	cids := pa.Value.(*types.ClusterList)
	if len(*cids) == 0 {
		return 0
	}

	attrFlags := uint8(0)
	attrFlags = setOptional(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(ClusterListAttr)

	length := uint8(ClusterIDLen * len(*cids))
	buf.WriteByte(length)

	for _, cid := range *cids {
		buf.Write(convert.Uint32Byte(cid))
	}

	return length + 3
}

func (pa *PathAttribute) serializeUnknownAttribute(buf *bytes.Buffer) uint16 {
	attrFlags := uint8(0)
	if pa.Optional {
		attrFlags = setOptional(attrFlags)
	}
	if pa.ExtendedLength {
		attrFlags = setExtendedLength(attrFlags)
	}
	attrFlags = setTransitive(attrFlags)

	buf.WriteByte(attrFlags)
	buf.WriteByte(pa.TypeCode)

	b := pa.Value.([]byte)
	if pa.ExtendedLength {
		l := len(b)
		buf.WriteByte(uint8(l >> 8))
		buf.WriteByte(uint8(l & 0x0000FFFF))
	} else {
		buf.WriteByte(uint8(len(b)))
	}
	buf.Write(b)

	return uint16(len(b)) + 3
}

func (pa *PathAttribute) serializeMultiProtocolReachNLRI(buf *bytes.Buffer, opt *EncodeOptions) uint16 {
	v := pa.Value.(MultiProtocolReachNLRI)
	pa.Optional = true

	tempBuf := bytes.NewBuffer(nil)
	v.serialize(tempBuf, opt)

	return pa.serializeGeneric(tempBuf.Bytes(), buf)
}

func (pa *PathAttribute) serializeMultiProtocolUnreachNLRI(buf *bytes.Buffer, opt *EncodeOptions) uint16 {
	v := pa.Value.(MultiProtocolUnreachNLRI)
	pa.Optional = true

	tempBuf := bytes.NewBuffer(nil)
	v.serialize(tempBuf, opt)

	return pa.serializeGeneric(tempBuf.Bytes(), buf)
}

func (pa *PathAttribute) serializeGeneric(b []byte, buf *bytes.Buffer) uint16 {
	attrFlags := uint8(0)
	if pa.Optional {
		attrFlags = setOptional(attrFlags)
	}
	if pa.Transitive {
		attrFlags = setTransitive(attrFlags)
	}

	if len(b) > math.MaxUint8 {
		pa.ExtendedLength = true
	}

	if pa.ExtendedLength {
		attrFlags = setExtendedLength(attrFlags)
	}

	if pa.Transitive {
		attrFlags = setTransitive(attrFlags)
	}

	buf.WriteByte(attrFlags)
	buf.WriteByte(pa.TypeCode)

	if pa.ExtendedLength {
		l := len(b)
		buf.WriteByte(uint8(l >> 8))
		buf.WriteByte(uint8(l & 0x0000FFFF))
	} else {
		buf.WriteByte(uint8(len(b)))
	}
	buf.Write(b)

	return uint16(len(b) + 2)
}

func fourBytesToUint32(address [4]byte) uint32 {
	return uint32(address[0])<<24 + uint32(address[1])<<16 + uint32(address[2])<<8 + uint32(address[3])
}

func read4BytesAsUint32(buf *bytes.Buffer) (uint32, error) {
	b := [4]byte{}
	n, err := buf.Read(b[:])
	if err != nil {
		return 0, err
	}
	if n != 4 {
		return 0, fmt.Errorf("unable to read as uint32. Expected 4 bytes but got only %d", n)
	}

	return fourBytesToUint32(b), nil
}

// AddOptionalPathAttributes adds optional path attributes to linked list pa
func (pa *PathAttribute) AddOptionalPathAttributes(p *route.Path) *PathAttribute {
	current := pa

	if p.BGPPath.Communities != nil && len(*p.BGPPath.Communities) > 0 {
		communities := &PathAttribute{
			TypeCode: CommunitiesAttr,
			Value:    p.BGPPath.Communities,
		}
		current.Next = communities
		current = communities
	}

	if p.BGPPath.LargeCommunities != nil && len(*p.BGPPath.LargeCommunities) > 0 {
		largeCommunities := &PathAttribute{
			TypeCode: LargeCommunitiesAttr,
			Value:    p.BGPPath.LargeCommunities,
		}
		current.Next = largeCommunities
		current = largeCommunities
	}

	return current
}

// PathAttributes converts a path object into a linked list of path attributes
func PathAttributes(p *route.Path, iBGP bool, rrClient bool) (*PathAttribute, error) {
	asPath := &PathAttribute{
		TypeCode: ASPathAttr,
		Value:    p.BGPPath.ASPath,
	}
	last := asPath

	origin := &PathAttribute{
		TypeCode: OriginAttr,
		Value:    p.BGPPath.BGPPathA.Origin,
	}
	last.Next = origin
	last = origin

	nextHop := &PathAttribute{
		TypeCode: NextHopAttr,
		Value:    p.BGPPath.BGPPathA.NextHop,
	}
	last.Next = nextHop
	last = nextHop

	if p.BGPPath.BGPPathA.MED != 0 {
		med := &PathAttribute{
			TypeCode: MEDAttr,
			Value:    p.BGPPath.BGPPathA.MED,
			Optional: true,
		}

		last.Next = med
		last = med
	}

	if p.BGPPath.BGPPathA.AtomicAggregate {
		atomicAggr := &PathAttribute{
			TypeCode: AtomicAggrAttr,
		}
		last.Next = atomicAggr
		last = atomicAggr
	}

	if p.BGPPath.BGPPathA.Aggregator != nil {
		aggregator := &PathAttribute{
			TypeCode: AggregatorAttr,
			Value:    *p.BGPPath.BGPPathA.Aggregator,
		}
		last.Next = aggregator
		last = aggregator
	}

	if iBGP {
		localPref := &PathAttribute{
			TypeCode: LocalPrefAttr,
			Value:    p.BGPPath.BGPPathA.LocalPref,
		}
		last.Next = localPref
		last = localPref
	}

	if rrClient {
		originatorID := &PathAttribute{
			TypeCode: OriginatorIDAttr,
			Value:    p.BGPPath.BGPPathA.OriginatorID,
		}
		last.Next = originatorID
		last = originatorID

		clusterList := &PathAttribute{
			TypeCode: ClusterListAttr,
			Value:    p.BGPPath.ClusterList,
		}
		last.Next = clusterList
		last = clusterList
	}

	optionals := last.AddOptionalPathAttributes(p)

	last = optionals
	for _, unknownAttr := range p.BGPPath.UnknownAttributes {
		last.Next = &PathAttribute{
			TypeCode:   unknownAttr.TypeCode,
			Optional:   unknownAttr.Optional,
			Transitive: unknownAttr.Transitive,
			Partial:    unknownAttr.Partial,
			Value:      unknownAttr.Value,
		}
		last = last.Next
	}

	return asPath, nil
}
