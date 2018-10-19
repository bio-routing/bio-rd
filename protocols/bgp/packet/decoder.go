package packet

import (
	"bytes"
	"fmt"
	"net"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/taktv6/tflow2/convert"
)

// Decode decodes a BGP message
func Decode(buf *bytes.Buffer, opt *DecodeOptions) (*BGPMessage, error) {
	hdr, err := decodeHeader(buf)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode header: %v", err)
	}

	body, err := decodeMsgBody(buf, hdr.Type, hdr.Length-MinLen, opt)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode message: %v", err)
	}

	return &BGPMessage{
		Header: hdr,
		Body:   body,
	}, nil
}

func decodeMsgBody(buf *bytes.Buffer, msgType uint8, l uint16, opt *DecodeOptions) (interface{}, error) {
	switch msgType {
	case OpenMsg:
		return DecodeOpenMsg(buf)
	case UpdateMsg:
		return decodeUpdateMsg(buf, l, opt)
	case KeepaliveMsg:
		return nil, nil // Nothing to decode in Keepalive message
	case NotificationMsg:
		return decodeNotificationMsg(buf)
	}
	return nil, fmt.Errorf("Unknown message type: %d", msgType)
}

func decodeUpdateMsg(buf *bytes.Buffer, l uint16, opt *DecodeOptions) (*BGPUpdate, error) {
	msg := &BGPUpdate{}

	err := decode.Decode(buf, []interface{}{&msg.WithdrawnRoutesLen})
	if err != nil {
		return msg, err
	}

	msg.WithdrawnRoutes, err = decodeNLRIs(buf, uint16(msg.WithdrawnRoutesLen), IPv4AFI, opt.AddPath)
	if err != nil {
		return msg, err
	}

	err = decode.Decode(buf, []interface{}{&msg.TotalPathAttrLen})
	if err != nil {
		return msg, err
	}

	msg.PathAttributes, err = decodePathAttrs(buf, msg.TotalPathAttrLen, opt)
	if err != nil {
		return msg, err
	}

	nlriLen := uint16(l) - 4 - uint16(msg.TotalPathAttrLen) - uint16(msg.WithdrawnRoutesLen)
	if nlriLen > 0 {
		msg.NLRI, err = decodeNLRIs(buf, nlriLen, IPv4AFI, opt.AddPath)
		if err != nil {
			return msg, err
		}
	}

	return msg, nil
}

func decodeNotificationMsg(buf *bytes.Buffer) (*BGPNotification, error) {
	msg := &BGPNotification{}

	fields := []interface{}{
		&msg.ErrorCode,
		&msg.ErrorSubcode,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return msg, err
	}

	if msg.ErrorCode > Cease {
		return msg, fmt.Errorf("Invalid error code: %d", msg.ErrorSubcode)
	}

	switch msg.ErrorCode {
	case MessageHeaderError:
		if msg.ErrorSubcode > BadMessageType || msg.ErrorSubcode == 0 {
			return invalidErrCode(msg)
		}
	case OpenMessageError:
		if msg.ErrorSubcode > UnacceptableHoldTime || msg.ErrorSubcode == 0 || msg.ErrorSubcode == DeprecatedOpenMsgError5 {
			return invalidErrCode(msg)
		}
	case UpdateMessageError:
		if msg.ErrorSubcode > MalformedASPath || msg.ErrorSubcode == 0 || msg.ErrorSubcode == DeprecatedUpdateMsgError7 {
			return invalidErrCode(msg)
		}
	case HoldTimeExpired:
		if msg.ErrorSubcode != 0 {
			return invalidErrCode(msg)
		}
	case FiniteStateMachineError:
		if msg.ErrorSubcode != 0 {
			return invalidErrCode(msg)
		}
	case Cease:
		if !(msg.ErrorSubcode == 0 || msg.ErrorSubcode == AdministrativeShutdown || msg.ErrorSubcode == AdministrativeReset) {
			return invalidErrCode(msg)
		}
	default:
		return invalidErrCode(msg)
	}

	return msg, nil
}

func invalidErrCode(n *BGPNotification) (*BGPNotification, error) {
	return n, fmt.Errorf("Invalid error sub code: %d/%d", n.ErrorCode, n.ErrorSubcode)
}

// DecodeOpenMsg decodes a BGP OPEN message
func DecodeOpenMsg(buf *bytes.Buffer) (*BGPOpen, error) {
	msg, err := _decodeOpenMsg(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode OPEN message: %v", err)
	}
	return msg.(*BGPOpen), err
}

func _decodeOpenMsg(buf *bytes.Buffer) (interface{}, error) {
	msg := &BGPOpen{}

	fields := []interface{}{
		&msg.Version,
		&msg.ASN,
		&msg.HoldTime,
		&msg.BGPIdentifier,
		&msg.OptParmLen,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return msg, err
	}

	err = validateOpen(msg)
	if err != nil {
		return nil, err
	}

	msg.OptParams, err = decodeOptParams(buf, msg.OptParmLen)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode optional parameters: %v", err)
	}

	return msg, nil
}

func decodeOptParams(buf *bytes.Buffer, optParmLen uint8) ([]OptParam, error) {
	optParams := make([]OptParam, 0)
	read := uint8(0)
	for read < optParmLen {
		o := OptParam{}
		fields := []interface{}{
			&o.Type,
			&o.Length,
		}

		err := decode.Decode(buf, fields)
		if err != nil {
			return nil, err
		}

		read += 2

		switch o.Type {
		case CapabilitiesParamType:
			caps, err := decodeCapabilities(buf, o.Length)
			if err != nil {
				return nil, fmt.Errorf("Unable to decode capabilities: %v", err)
			}

			o.Value = caps
			optParams = append(optParams, o)
			for _, cap := range caps {
				read += cap.Length + 2
			}
		default:
			return nil, fmt.Errorf("Unrecognized option: %d", o.Type)
		}

	}

	return optParams, nil
}

func decodeCapabilities(buf *bytes.Buffer, length uint8) (Capabilities, error) {
	ret := make(Capabilities, 0)
	read := uint8(0)
	for read < length {
		cap, err := decodeCapability(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode capability: %v", err)
		}

		ret = append(ret, cap)
		read += cap.Length + 2
	}

	return ret, nil
}

func decodeCapability(buf *bytes.Buffer) (Capability, error) {
	cap := Capability{}
	fields := []interface{}{
		&cap.Code,
		&cap.Length,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return cap, err
	}

	switch cap.Code {
	case MultiProtocolCapabilityCode:
		mpCap, err := decodeMultiProtocolCapability(buf)
		if err != nil {
			return cap, fmt.Errorf("Unable to decode multi protocol capability")
		}
		cap.Value = mpCap
	case AddPathCapabilityCode:
		addPathCap, err := decodeAddPathCapability(buf)
		if err != nil {
			return cap, fmt.Errorf("Unable to decode add path capability: %v", err)
		}
		cap.Value = addPathCap
	case ASN4CapabilityCode:
		asn4Cap, err := decodeASN4Capability(buf)
		if err != nil {
			return cap, fmt.Errorf("Unable to decode 4 octet ASN capability: %v", err)
		}
		cap.Value = asn4Cap
	default:
		for i := uint8(0); i < cap.Length; i++ {
			_, err := buf.ReadByte()
			if err != nil {
				return cap, fmt.Errorf("Read failed: %v", err)
			}
		}
	}

	return cap, nil
}

func decodeMultiProtocolCapability(buf *bytes.Buffer) (MultiProtocolCapability, error) {
	mpCap := MultiProtocolCapability{}
	reserved := uint8(0)
	fields := []interface{}{
		&mpCap.AFI, &reserved, &mpCap.SAFI,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return mpCap, err
	}

	return mpCap, nil
}

func decodeAddPathCapability(buf *bytes.Buffer) (AddPathCapability, error) {
	addPathCap := AddPathCapability{}
	fields := []interface{}{
		&addPathCap.AFI,
		&addPathCap.SAFI,
		&addPathCap.SendReceive,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return addPathCap, err
	}

	return addPathCap, nil
}

func decodeASN4Capability(buf *bytes.Buffer) (ASN4Capability, error) {
	asn4Cap := ASN4Capability{}
	fields := []interface{}{
		&asn4Cap.ASN4,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return asn4Cap, err
	}

	return asn4Cap, nil
}

func validateOpen(msg *BGPOpen) error {
	if msg.Version != BGP4Version {
		return BGPError{
			ErrorCode:    OpenMessageError,
			ErrorSubCode: UnsupportedVersionNumber,
			ErrorStr:     fmt.Sprintf("Unsupported version number"),
		}
	}
	if !isValidIdentifier(msg.BGPIdentifier) {
		return BGPError{
			ErrorCode:    OpenMessageError,
			ErrorSubCode: BadBGPIdentifier,
			ErrorStr:     fmt.Sprintf("Invalid BGP identifier"),
		}
	}

	return nil
}

func isValidIdentifier(id uint32) bool {
	addr := net.IP(convert.Uint32Byte(id))
	if addr.IsLoopback() {
		return false
	}

	if addr.IsMulticast() {
		return false
	}

	if addr[0] == 0 {
		return false
	}

	if addr[0] == 255 && addr[1] == 255 && addr[2] == 255 && addr[3] == 255 {
		return false
	}

	return true
}

func decodeHeader(buf *bytes.Buffer) (*BGPHeader, error) {
	hdr := &BGPHeader{}

	marker := make([]byte, MarkerLen)
	n, err := buf.Read(marker)
	if err != nil {
		return hdr, BGPError{
			ErrorCode:    Cease,
			ErrorSubCode: 0,
			ErrorStr:     fmt.Sprintf("Failed to read from buffer: %v", err.Error()),
		}
	}

	if n != MarkerLen {
		return hdr, BGPError{
			ErrorCode:    Cease,
			ErrorSubCode: 0,
			ErrorStr:     fmt.Sprintf("Unable to read marker"),
		}
	}

	for i := range marker {
		if marker[i] != 255 {
			return nil, BGPError{
				ErrorCode:    MessageHeaderError,
				ErrorSubCode: ConnectionNotSync,
				ErrorStr:     fmt.Sprintf("Invalid marker: %v", marker),
			}
		}
	}

	fields := []interface{}{
		&hdr.Length,
		&hdr.Type,
	}

	err = decode.Decode(buf, fields)
	if err != nil {
		return hdr, BGPError{
			ErrorCode:    Cease,
			ErrorSubCode: 0,
			ErrorStr:     fmt.Sprintf("%v", err.Error()),
		}
	}

	if hdr.Length < MinLen || hdr.Length > MaxLen {
		return hdr, BGPError{
			ErrorCode:    MessageHeaderError,
			ErrorSubCode: BadMessageLength,
			ErrorStr:     fmt.Sprintf("Invalid length in BGP header: %v", hdr.Length),
		}
	}

	if hdr.Type > KeepaliveMsg || hdr.Type == 0 {
		return hdr, BGPError{
			ErrorCode:    MessageHeaderError,
			ErrorSubCode: BadMessageType,
			ErrorStr:     fmt.Sprintf("Invalid message type: %d", hdr.Type),
		}
	}

	return hdr, nil
}
