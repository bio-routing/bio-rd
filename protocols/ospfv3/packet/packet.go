package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"
)

type OSPFv3Message struct {
	Version      uint8
	Type         uint8
	PacketLength uint16
	RouterID     ID
	AreaID       ID
	Checksum     uint16
	InstanceID   uint8
	Body         Serializable
}

const OSPFv3MessageHeaderLength = 16

func (x *OSPFv3Message) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(x.Version)
	buf.WriteByte(x.Type)
	buf.Write(convert.Uint16Byte(x.PacketLength))
	x.RouterID.Serialize(buf)
	x.AreaID.Serialize(buf)
	buf.Write(convert.Uint16Byte(x.Checksum))
	buf.WriteByte(x.InstanceID)
	buf.WriteByte(0) // 1 byte reserved
	x.Body.Serialize(buf)
}

func DeserializeOSPFv3Message(buf *bytes.Buffer) (*OSPFv3Message, int, error) {
	pdu := &OSPFv3Message{}

	var readBytes int
	var err error
	var fields []interface{}

	fields = []interface{}{
		&pdu.Version,
		&pdu.Type,
		&pdu.PacketLength,
		&pdu.RouterID,
		&pdu.AreaID,
		&pdu.Checksum,
		&pdu.InstanceID,
		new(uint8), // 1 byte reserved
	}

	err = decode.Decode(buf, fields)
	if err != nil {
		return nil, readBytes, fmt.Errorf("Unable to decode fields: %v", err)
	}
	readBytes += 16

	n, err := pdu.ReadBody(buf)
	if err != nil {
		return nil, readBytes, errors.Wrap(err, "unable to decode message body")
	}
	readBytes += n

	return pdu, readBytes, nil
}

func (m *OSPFv3Message) ReadBody(buf *bytes.Buffer) (int, error) {
	bodyLength := m.PacketLength - OSPFv3MessageHeaderLength
	var body Serializable
	var readBytes int
	var err error

	switch m.Type {
	case 1:
		body, readBytes, err = DeserializeHello(buf, bodyLength)
	case 2:
		body, readBytes, err = DeserializeDatabaseDescription(buf, bodyLength)
	case 3:
		body, readBytes, err = DeserializeLinkStateRequestMsg(buf, bodyLength)
	case 4:
		body, readBytes, err = DeserializeLinkStateUpdate(buf)
	case 5:
		body, readBytes, err = DeserializeLinkStateAcknowledgement(buf, bodyLength)
	default:
		return 0, fmt.Errorf("unknown message type: %d", m.Type)
	}

	if err != nil {
		return 0, err
	}

	m.Body = body
	return readBytes, nil
}

type Hello struct {
	InterfaceID              ID
	RouterPriority           uint8
	Options                  RouterOptions
	HelloInterval            uint16
	RouterDeadInterval       uint16
	DesignatedRouterID       ID
	BackupDesignatedRouterID ID
	Neighbors                []ID
}

func (x *Hello) Serialize(buf *bytes.Buffer) {
	x.InterfaceID.Serialize(buf)
	buf.WriteByte(x.RouterPriority)
	x.Options.Serialize(buf)
	buf.Write(convert.Uint16Byte(x.HelloInterval))
	buf.Write(convert.Uint16Byte(x.RouterDeadInterval))
	x.DesignatedRouterID.Serialize(buf)
	x.BackupDesignatedRouterID.Serialize(buf)
	for i := range x.Neighbors {
		x.Neighbors[i].Serialize(buf)
	}
}

func DeserializeHello(buf *bytes.Buffer, bodyLength uint16) (*Hello, int, error) {
	pdu := &Hello{}

	var readBytes int
	var err error
	var fields []interface{}

	fields = []interface{}{
		&pdu.InterfaceID,
		&pdu.RouterPriority,
		&pdu.Options,
		&pdu.HelloInterval,
		&pdu.RouterDeadInterval,
		&pdu.DesignatedRouterID,
		&pdu.BackupDesignatedRouterID,
	}

	err = decode.Decode(buf, fields)
	if err != nil {
		return nil, readBytes, fmt.Errorf("Unable to decode fields: %v", err)
	}
	readBytes += 20

	for i := readBytes; i < int(bodyLength); {
		id, n, err := DeserializeID(buf)
		if err != nil {
			return nil, readBytes, errors.Wrap(err, "unable to decode neighbor id")
		}
		pdu.Neighbors = append(pdu.Neighbors, id)
		i += n
		readBytes += n
	}

	return pdu, readBytes, nil
}

type DatabaseDescription struct {
	Options          RouterOptions
	InterfaceMTU     uint16
	DBFlags          uint8
	DDSequenceNumber uint32
	LSAHeaders       []*LSA
}

func (x *DatabaseDescription) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(0) // 1 byte reserved
	x.Options.Serialize(buf)
	buf.Write(convert.Uint16Byte(x.InterfaceMTU))
	buf.WriteByte(0) // 1 byte reserved
	buf.WriteByte(x.DBFlags)
	buf.Write(convert.Uint32Byte(x.DDSequenceNumber))
	for i := range x.LSAHeaders {
		x.LSAHeaders[i].Serialize(buf, true)
	}
}

func DeserializeDatabaseDescription(buf *bytes.Buffer, bodyLength uint16) (*DatabaseDescription, int, error) {
	pdu := &DatabaseDescription{}

	var readBytes int
	var err error
	var fields []interface{}

	fields = []interface{}{
		new(uint8),
		&pdu.Options,
		&pdu.InterfaceMTU,
		new(uint8),
		&pdu.DBFlags,
		&pdu.DDSequenceNumber,
	}

	err = decode.Decode(buf, fields)
	if err != nil {
		return nil, readBytes, fmt.Errorf("Unable to decode fields: %v", err)
	}
	readBytes += 12

	for i := readBytes; i < int(bodyLength); {
		tlv, n, err := DeserializeLSA(buf, true)
		if err != nil {
			return nil, 0, errors.Wrap(err, "Unable to decode")
		}
		pdu.LSAHeaders = append(pdu.LSAHeaders, tlv)
		i += n
		readBytes += n
	}

	return pdu, readBytes, nil
}

type LinkStateRequestMsg struct {
	Requests []LinkStateRequest
}

func (x *LinkStateRequestMsg) Serialize(buf *bytes.Buffer) {
	for i := range x.Requests {
		x.Requests[i].Serialize(buf)
	}
}

func DeserializeLinkStateRequestMsg(buf *bytes.Buffer, bodyLength uint16) (*LinkStateRequestMsg, int, error) {
	pdu := &LinkStateRequestMsg{}

	var readBytes int
	for readBytes < int(bodyLength) {
		req, n, err := DeserializeLinkStateRequest(buf)
		if err != nil {
			return nil, readBytes, errors.Wrap(err, "unable to decode LinkStateRequest")
		}
		pdu.Requests = append(pdu.Requests, req)
		readBytes += n
	}

	return pdu, readBytes, nil
}

type LinkStateRequest struct {
	LSType            LSType
	LinkStateID       ID
	AdvertisingRouter ID
}

func (x *LinkStateRequest) Serialize(buf *bytes.Buffer) {
	buf.Write([]byte{0, 0}) // 2 bytes reserved
	x.LSType.Serialize(buf)
	x.LinkStateID.Serialize(buf)
	x.AdvertisingRouter.Serialize(buf)
}

func DeserializeLinkStateRequest(buf *bytes.Buffer) (LinkStateRequest, int, error) {
	pdu := LinkStateRequest{}

	var readBytes int
	var err error
	var fields []interface{}

	fields = []interface{}{
		new(uint16), // 2 bytes reserved
		&pdu.LSType,
		&pdu.LinkStateID,
		&pdu.AdvertisingRouter,
	}

	err = decode.Decode(buf, fields)
	if err != nil {
		return pdu, readBytes, fmt.Errorf("Unable to decode fields: %v", err)
	}
	readBytes += 12

	return pdu, readBytes, nil
}

type LinkStateUpdate struct {
	LSAs []*LSA
}

func (x *LinkStateUpdate) Serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint32Byte(uint32(len(x.LSAs))))
	for i := range x.LSAs {
		x.LSAs[i].Serialize(buf, false)
	}
}

func DeserializeLinkStateUpdate(buf *bytes.Buffer) (*LinkStateUpdate, int, error) {
	pdu := &LinkStateUpdate{}

	var lsaCount uint32
	if err := binary.Read(buf, binary.BigEndian, &lsaCount); err != nil {
		return nil, 0, errors.Wrap(err, "unable to decode LSA count")
	}
	readBytes := 4

	for i := 0; i < int(lsaCount); i++ {
		tlv, n, err := DeserializeLSA(buf, false)
		if err != nil {
			return nil, 0, errors.Wrap(err, "unable to decode LSA")
		}
		pdu.LSAs = append(pdu.LSAs, tlv)
		readBytes += n
	}

	return pdu, readBytes, nil
}

type LinkStateAcknowledgement struct {
	LSAHeaders []*LSA
}

func (x *LinkStateAcknowledgement) Serialize(buf *bytes.Buffer) {
	for i := range x.LSAHeaders {
		x.LSAHeaders[i].Serialize(buf, true)
	}
}

func DeserializeLinkStateAcknowledgement(buf *bytes.Buffer, bodyLength uint16) (*LinkStateAcknowledgement, int, error) {
	pdu := &LinkStateAcknowledgement{}

	var readBytes int

	for i := 0; i < int(bodyLength); {
		tlv, n, err := DeserializeLSA(buf, true)
		if err != nil {
			return nil, 0, errors.Wrap(err, "Unable to decode")
		}
		pdu.LSAHeaders = append(pdu.LSAHeaders, tlv)
		i += n
		readBytes += n
	}

	return pdu, readBytes, nil
}
