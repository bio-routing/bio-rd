package packet

import (
	"bytes"
	"fmt"

	"github.com/taktv6/tflow2/convert"
)

func SerializeKeepaliveMsg() []byte {
	keepaliveLen := uint16(19)
	buf := bytes.NewBuffer(make([]byte, 0, keepaliveLen))
	serializeHeader(buf, keepaliveLen, KeepaliveMsg)

	return buf.Bytes()
}

func SerializeNotificationMsg(msg *BGPNotification) []byte {
	notificationLen := uint16(21)
	buf := bytes.NewBuffer(make([]byte, 0, notificationLen))
	serializeHeader(buf, notificationLen, NotificationMsg)
	buf.WriteByte(msg.ErrorCode)
	buf.WriteByte(msg.ErrorSubcode)

	return buf.Bytes()
}

func SerializeOpenMsg(msg *BGPOpen) []byte {
	optParmsBuf := bytes.NewBuffer(make([]byte, 0))
	serializeOptParams(optParmsBuf, msg.OptParams)
	optParms := optParmsBuf.Bytes()
	openLen := uint16(len(optParms) + MinOpenLen)

	buf := bytes.NewBuffer(make([]byte, 0, openLen))
	serializeHeader(buf, openLen, OpenMsg)

	buf.WriteByte(msg.Version)
	buf.Write(convert.Uint16Byte(msg.ASN))
	buf.Write(convert.Uint16Byte(msg.HoldTime))
	buf.Write(convert.Uint32Byte(msg.BGPIdentifier))

	buf.WriteByte(uint8(len(optParms)))
	buf.Write(optParms)

	return buf.Bytes()
}

func serializeOptParams(buf *bytes.Buffer, params []OptParam) {
	for _, param := range params {
		tmpBuf := bytes.NewBuffer(make([]byte, 0))
		param.Value.serialize(tmpBuf)
		payload := tmpBuf.Bytes()

		buf.WriteByte(param.Type)
		buf.WriteByte(uint8(len(payload)))
		buf.Write(payload)
	}
}

func serializeHeader(buf *bytes.Buffer, length uint16, typ uint8) {
	buf.Write([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	buf.Write(convert.Uint16Byte(length))
	buf.WriteByte(typ)
}

func (b *BGPUpdateAddPath) SerializeUpdate(opt *Options) ([]byte, error) {
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
		paLen := int(pa.serialize(pathAttributesBuf, opt))
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

func (b *BGPUpdate) SerializeUpdate(opt *Options) ([]byte, error) {
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
		paLen := int(pa.serialize(pathAttributesBuf, opt))
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
