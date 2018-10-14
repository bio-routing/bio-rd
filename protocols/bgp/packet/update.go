package packet

import (
	"bytes"
	"fmt"

	"github.com/taktv6/tflow2/convert"
)

// BGPUpdate represents a BGP Update message
type BGPUpdate struct {
	WithdrawnRoutesLen uint16
	WithdrawnRoutes    *NLRI
	TotalPathAttrLen   uint16
	PathAttributes     *PathAttribute
	NLRI               *NLRI
}

// DecodeUpdateMsg decodes a BGP Update Message
func DecodeUpdateMsg(buf *bytes.Buffer, l uint16, opt *DecodeOptions) (*BGPUpdate, error) {
	msg := &BGPUpdate{}

	err := decode(buf, []interface{}{&msg.WithdrawnRoutesLen})
	if err != nil {
		return msg, err
	}

	msg.WithdrawnRoutes, err = decodeNLRIs(buf, uint16(msg.WithdrawnRoutesLen), opt)
	if err != nil {
		return msg, err
	}

	err = decode(buf, []interface{}{&msg.TotalPathAttrLen})
	if err != nil {
		return msg, err
	}

	msg.PathAttributes, err = decodePathAttrs(buf, msg.TotalPathAttrLen, opt)
	if err != nil {
		return msg, err
	}

	nlriLen := uint16(l) - 4 - uint16(msg.TotalPathAttrLen) - uint16(msg.WithdrawnRoutesLen)
	if nlriLen > 0 {
		msg.NLRI, err = decodeNLRIs(buf, nlriLen, opt)
		if err != nil {
			return msg, err
		}
	}

	return msg, nil
}

// SerializeUpdate serializes an BGPUpdate to wire format
func (b *BGPUpdate) SerializeUpdate(opt *EncodeOptions) ([]byte, error) {
	budget := MaxLen - MinLen
	nlriLen := 0
	buf := bytes.NewBuffer(nil)

	withdrawBuf := bytes.NewBuffer(nil)
	for withdraw := b.WithdrawnRoutes; withdraw != nil; withdraw = withdraw.Next {
		if opt.UseAddPath {
			nlriLen = int(withdraw.serializeAddPath(withdrawBuf))
		} else {
			nlriLen = int(withdraw.serialize(withdrawBuf))
		}

		budget -= nlriLen
		if budget < 0 {
			return nil, fmt.Errorf("update too long")
		}
	}

	pathAttributesBuf := bytes.NewBuffer(nil)
	for pa := b.PathAttributes; pa != nil; pa = pa.Next {
		paLen := int(pa.Serialize(pathAttributesBuf, opt))
		budget -= paLen
		if budget < 0 {
			return nil, fmt.Errorf("update too long")
		}
	}

	nlriBuf := bytes.NewBuffer(nil)
	for nlri := b.NLRI; nlri != nil; nlri = nlri.Next {
		if opt.UseAddPath {
			nlriLen = int(nlri.serializeAddPath(nlriBuf))
		} else {
			nlriLen = int(nlri.serialize(nlriBuf))
		}

		budget -= nlriLen
		if budget < 0 {
			return nil, fmt.Errorf("update too long")
		}
	}

	withdrawnRoutesLen := withdrawBuf.Len()
	if withdrawnRoutesLen > 65535 {
		return nil, fmt.Errorf("Invalid Withdrawn Routes Length: %d", withdrawnRoutesLen)
	}

	totalPathAttributesLen := pathAttributesBuf.Len()
	if totalPathAttributesLen > 65535 {
		return nil, fmt.Errorf("Invalid Total Path Attribute Length: %d", totalPathAttributesLen)
	}

	totalLength := 2 + withdrawnRoutesLen + totalPathAttributesLen + 2 + nlriBuf.Len() + 19
	if totalLength > 4096 {
		return nil, fmt.Errorf("Update too long: %d bytes", totalLength)
	}

	serializeHeader(buf, uint16(totalLength), UpdateMsg)

	buf.Write(convert.Uint16Byte(uint16(withdrawnRoutesLen)))
	buf.Write(withdrawBuf.Bytes())

	buf.Write(convert.Uint16Byte(uint16(totalPathAttributesLen)))
	buf.Write(pathAttributesBuf.Bytes())

	buf.Write(nlriBuf.Bytes())

	return buf.Bytes(), nil
}

// SerializeUpdateAddPath serializes an BGPUpdate with add-path to wire format
func (b *BGPUpdate) SerializeUpdateAddPath(opt *EncodeOptions) ([]byte, error) {
	budget := MaxLen - MinLen
	buf := bytes.NewBuffer(nil)

	withdrawBuf := bytes.NewBuffer(nil)
	for withdraw := b.WithdrawnRoutes; withdraw != nil; withdraw = withdraw.Next {
		nlriLen := int(withdraw.serialize(withdrawBuf))
		budget -= nlriLen
		if budget < 0 {
			return nil, fmt.Errorf("update too long")
		}
	}

	pathAttributesBuf := bytes.NewBuffer(nil)
	for pa := b.PathAttributes; pa != nil; pa = pa.Next {
		paLen := int(pa.Serialize(pathAttributesBuf, opt))
		budget -= paLen
		if budget < 0 {
			return nil, fmt.Errorf("update too long")
		}
	}

	nlriBuf := bytes.NewBuffer(nil)
	for nlri := b.NLRI; nlri != nil; nlri = nlri.Next {
		nlriLen := int(nlri.serialize(nlriBuf))
		budget -= nlriLen
		if budget < 0 {
			return nil, fmt.Errorf("update too long")
		}
	}

	withdrawnRoutesLen := withdrawBuf.Len()
	if withdrawnRoutesLen > 65535 {
		return nil, fmt.Errorf("Invalid Withdrawn Routes Length: %d", withdrawnRoutesLen)
	}

	totalPathAttributesLen := pathAttributesBuf.Len()
	if totalPathAttributesLen > 65535 {
		return nil, fmt.Errorf("Invalid Total Path Attribute Length: %d", totalPathAttributesLen)
	}

	totalLength := 2 + withdrawnRoutesLen + totalPathAttributesLen + 2 + nlriBuf.Len() + 19
	if totalLength > 4096 {
		return nil, fmt.Errorf("Update too long: %d bytes", totalLength)
	}

	serializeHeader(buf, uint16(totalLength), UpdateMsg)

	buf.Write(convert.Uint16Byte(uint16(withdrawnRoutesLen)))
	buf.Write(withdrawBuf.Bytes())

	buf.Write(convert.Uint16Byte(uint16(totalPathAttributesLen)))
	buf.Write(pathAttributesBuf.Bytes())

	buf.Write(nlriBuf.Bytes())

	return buf.Bytes(), nil
}

// AddressFamily gets the address family of an update
func (b *BGPUpdate) AddressFamily() (afi uint16, safi uint8) {
	if b.WithdrawnRoutes != nil || b.NLRI != nil {
		return IPv4AFI, UnicastSAFI
	}

	for cur := b.PathAttributes; cur != nil; cur = cur.Next {
		if cur.TypeCode == MultiProtocolReachNLRICode {
			a := cur.Value.(MultiProtocolReachNLRI)
			return a.AFI, a.SAFI
		}

		if cur.TypeCode == MultiProtocolUnreachNLRICode {
			a := cur.Value.(MultiProtocolUnreachNLRI)
			return a.AFI, a.SAFI
		}
	}

	return
}
