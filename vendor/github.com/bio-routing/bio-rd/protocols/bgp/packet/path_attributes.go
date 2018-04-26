package packet

import (
	"bytes"
	"fmt"

	"github.com/taktv6/tflow2/convert"
)

func decodePathAttrs(buf *bytes.Buffer, tpal uint16) (*PathAttribute, error) {
	var ret *PathAttribute
	var eol *PathAttribute
	var pa *PathAttribute
	var err error
	var consumed uint16

	p := uint16(0)
	for p < tpal {
		pa, consumed, err = decodePathAttr(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode path attr: %v", err)
		}
		p += consumed

		if ret == nil {
			ret = pa
			eol = pa
		} else {
			eol.Next = pa
			eol = pa
		}
	}

	return ret, nil
}

func decodePathAttr(buf *bytes.Buffer) (pa *PathAttribute, consumed uint16, err error) {
	pa = &PathAttribute{}

	err = decodePathAttrFlags(buf, pa)
	if err != nil {
		return nil, consumed, fmt.Errorf("Unable to get path attribute flags: %v", err)
	}
	consumed++

	err = decode(buf, []interface{}{&pa.TypeCode})
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
			return nil, consumed, fmt.Errorf("Failed to decode Origin: %v", err)
		}
	case ASPathAttr:
		if err := pa.decodeASPath(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode AS Path: %v", err)
		}
	case NextHopAttr:
		if err := pa.decodeNextHop(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode Next-Hop: %v", err)
		}
	case MEDAttr:
		if err := pa.decodeMED(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode MED: %v", err)
		}
	case LocalPrefAttr:
		if err := pa.decodeLocalPref(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode local pref: %v", err)
		}
	case AggregatorAttr:
		if err := pa.decodeAggregator(buf); err != nil {
			return nil, consumed, fmt.Errorf("Failed to decode Aggregator: %v", err)
		}
	case AtomicAggrAttr:
		// Nothing to do for 0 octet long attribute
	default:
		return nil, consumed, fmt.Errorf("Invalid Attribute Type Code: %v", pa.TypeCode)
	}

	return pa, consumed + pa.Length, nil
}

func (pa *PathAttribute) decodeOrigin(buf *bytes.Buffer) error {
	origin := uint8(0)

	p := uint16(0)
	err := decode(buf, []interface{}{&origin})
	if err != nil {
		return fmt.Errorf("Unable to decode: %v", err)
	}

	pa.Value = origin
	p++

	return dumpNBytes(buf, pa.Length-p)
}

func (pa *PathAttribute) decodeASPath(buf *bytes.Buffer) error {
	pa.Value = make(ASPath, 0)

	p := uint16(0)
	for p < pa.Length {
		segment := ASPathSegment{
			ASNs: make([]uint32, 0),
		}

		err := decode(buf, []interface{}{&segment.Type, &segment.Count})
		if err != nil {
			return err
		}
		p += 2

		if segment.Type != ASSet && segment.Type != ASSequence {
			return fmt.Errorf("Invalid AS Path segment type: %d", segment.Type)
		}

		if segment.Count == 0 {
			return fmt.Errorf("Invalid AS Path segment length: %d", segment.Count)
		}

		for i := uint8(0); i < segment.Count; i++ {
			asn := uint16(0)

			err := decode(buf, []interface{}{&asn})
			if err != nil {
				return err
			}
			p += 2

			segment.ASNs = append(segment.ASNs, uint32(asn))
		}
		pa.Value = append(pa.Value.(ASPath), segment)
	}

	return nil
}

func (pa *PathAttribute) decodeNextHop(buf *bytes.Buffer) error {
	addr := [4]byte{}

	p := uint16(0)
	n, err := buf.Read(addr[:])
	if err != nil {
		return err
	}
	if n != 4 {
		return fmt.Errorf("Unable to read next hop: buf.Read read %d bytes", n)
	}

	pa.Value = addr
	p += 4

	return dumpNBytes(buf, pa.Length-p)
}

func (pa *PathAttribute) decodeMED(buf *bytes.Buffer) error {
	med, err := pa.decodeUint32(buf)
	if err != nil {
		return fmt.Errorf("Unable to recode local pref: %v", err)
	}

	pa.Value = uint32(med)
	return nil
}

func (pa *PathAttribute) decodeLocalPref(buf *bytes.Buffer) error {
	lpref, err := pa.decodeUint32(buf)
	if err != nil {
		return fmt.Errorf("Unable to recode local pref: %v", err)
	}

	pa.Value = uint32(lpref)
	return nil
}

func (pa *PathAttribute) decodeAggregator(buf *bytes.Buffer) error {
	aggr := Aggretator{}

	p := uint16(0)
	err := decode(buf, []interface{}{&aggr.ASN})
	if err != nil {
		return err
	}
	p += 2

	n, err := buf.Read(aggr.Addr[:])
	if err != nil {
		return err
	}
	if n != 4 {
		return fmt.Errorf("Unable to read aggregator IP: buf.Read read %d bytes", n)
	}
	p += 4

	pa.Value = aggr
	return dumpNBytes(buf, pa.Length-p)
}

func (pa *PathAttribute) setLength(buf *bytes.Buffer) (int, error) {
	bytesRead := 0
	if pa.ExtendedLength {
		err := decode(buf, []interface{}{&pa.Length})
		if err != nil {
			return 0, err
		}
		bytesRead = 2
	} else {
		x := uint8(0)
		err := decode(buf, []interface{}{&x})
		if err != nil {
			return 0, err
		}
		pa.Length = uint16(x)
		bytesRead = 1
	}
	return bytesRead, nil
}

func (pa *PathAttribute) decodeUint32(buf *bytes.Buffer) (uint32, error) {
	var v uint32

	p := uint16(0)
	err := decode(buf, []interface{}{&v})
	if err != nil {
		return 0, err
	}

	p += 4
	err = dumpNBytes(buf, pa.Length-p)
	if err != nil {
		return 0, fmt.Errorf("dumpNBytes failed: %v", err)
	}

	return v, nil
}

func (pa *PathAttribute) ASPathString() (ret string) {
	for _, p := range *pa.Value.(*ASPath) {
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

func (pa *PathAttribute) ASPathLen() (ret uint16) {
	for _, p := range *pa.Value.(*ASPath) {
		if p.Type == ASSet {
			ret++
			continue
		}
		ret += uint16(len(p.ASNs))
	}

	return
}

// dumpNBytes is used to dump n bytes of buf. This is useful in case an path attributes
// length doesn't match a fixed length's attributes length (e.g. ORIGIN is always an octet)
func dumpNBytes(buf *bytes.Buffer, n uint16) error {
	if n <= 0 {
		return nil
	}
	dump := make([]byte, n)
	err := decode(buf, []interface{}{&dump})
	if err != nil {
		return err
	}
	return nil
}

func (pa *PathAttribute) serialize(buf *bytes.Buffer) uint8 {
	pathAttrLen := uint8(0)

	switch pa.TypeCode {
	case OriginAttr:
		pathAttrLen = pa.serializeOrigin(buf)
	case ASPathAttr:
		pathAttrLen = pa.serializeASPath(buf)
	case NextHopAttr:
		pathAttrLen = pa.serializeNextHop(buf)
	case MEDAttr:
		pathAttrLen = pa.serializeMED(buf)
	case LocalPrefAttr:
		pathAttrLen = pa.serializeLocalpref(buf)
	case AtomicAggrAttr:
		pathAttrLen = pa.serializeAtomicAggregate(buf)
	case AggregatorAttr:
		pathAttrLen = pa.serializeAggregator(buf)
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
	return 4
}

func (pa *PathAttribute) serializeASPath(buf *bytes.Buffer) uint8 {
	attrFlags := uint8(0)
	attrFlags = setTransitive(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(ASPathAttr)
	length := uint8(2)
	asPath := pa.Value.(ASPath)
	for _, segment := range asPath {
		buf.WriteByte(segment.Type)
		buf.WriteByte(uint8(len(segment.ASNs)))
		for _, asn := range segment.ASNs {
			buf.Write(convert.Uint16Byte(uint16(asn)))
		}
		length += 2 + uint8(len(segment.ASNs))*2
	}

	return length
}

func (pa *PathAttribute) serializeNextHop(buf *bytes.Buffer) uint8 {
	attrFlags := uint8(0)
	attrFlags = setTransitive(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(NextHopAttr)
	length := uint8(4)
	buf.WriteByte(length)
	addr := pa.Value.([4]byte)
	buf.Write(addr[:])
	return 7
}

func (pa *PathAttribute) serializeMED(buf *bytes.Buffer) uint8 {
	attrFlags := uint8(0)
	attrFlags = setOptional(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(MEDAttr)
	length := uint8(4)
	buf.WriteByte(length)
	buf.Write(convert.Uint32Byte(pa.Value.(uint32)))
	return 7
}

func (pa *PathAttribute) serializeLocalpref(buf *bytes.Buffer) uint8 {
	attrFlags := uint8(0)
	attrFlags = setTransitive(attrFlags)
	buf.WriteByte(attrFlags)
	buf.WriteByte(LocalPrefAttr)
	length := uint8(4)
	buf.WriteByte(length)
	buf.Write(convert.Uint32Byte(pa.Value.(uint32)))
	return 7
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
	length := uint8(2)
	buf.WriteByte(length)
	buf.Write(convert.Uint16Byte(pa.Value.(uint16)))
	return 5
}
