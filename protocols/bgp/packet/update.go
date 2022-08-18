package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/tflow2/convert"
)

type BGPUpdate struct {
	WithdrawnRoutesLen uint16
	WithdrawnRoutes    *NLRI
	TotalPathAttrLen   uint16
	PathAttributes     *PathAttribute
	NLRI               *NLRI
	SAFI               uint8
}

// SerializeUpdate serializes an BGPUpdate to wire format
func (b *BGPUpdate) SerializeUpdate(opt *EncodeOptions) ([]byte, error) {
	budget := MaxLen - MinLen
	buf := bytes.NewBuffer(nil)

	withdrawBuf := bytes.NewBuffer(nil)
	for withdraw := b.WithdrawnRoutes; withdraw != nil; withdraw = withdraw.Next {
		budget -= int(withdraw.serialize(withdrawBuf, opt.UseAddPath, b.SAFI))
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
		budget -= int(nlri.serialize(nlriBuf, opt.UseAddPath, b.SAFI))
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

func (b *BGPUpdate) IsEndOfRIBMarker() bool {
	return b.WithdrawnRoutesLen == 0 && b.NLRI == nil
}
