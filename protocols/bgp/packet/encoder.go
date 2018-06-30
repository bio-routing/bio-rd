package packet

import (
	"bytes"

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
