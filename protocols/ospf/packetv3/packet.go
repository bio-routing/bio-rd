package packetv3

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/util/checksum"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"
)

const OSPFProtocolNumber = 89
const expectedVersion = 3

type OSPFMessageType uint8

// OSPF message types
const (
	MsgTypeUnknown OSPFMessageType = iota
	MsgTypeHello
	MsgTypeDatabaseDescription
	MsgTypeLinkStateRequest
	MsgTypeLinkStateUpdate
	MsgTypeLinkStateAcknowledgment
)

type OSPFv3Message struct {
	Version      uint8
	Type         OSPFMessageType
	PacketLength uint16
	RouterID     ID
	AreaID       ID
	Checksum     uint16
	InstanceID   uint8
	Body         Serializable
}

const OSPFv3MessageHeaderLength = 16
const OSPFv3MessagePacketLengthAtByte = 2
const OSPFv3MessageChecksumAtByte = 12

func (x *OSPFv3Message) Serialize(out *bytes.Buffer, src, dst net.IP) {
	buf := bytes.NewBuffer(nil)

	buf.WriteByte(x.Version)
	buf.WriteByte(uint8(x.Type))
	buf.Write(convert.Uint16Byte(x.PacketLength))
	x.RouterID.Serialize(buf)
	x.AreaID.Serialize(buf)
	buf.Write(convert.Uint16Byte(x.Checksum))
	buf.WriteByte(x.InstanceID)
	buf.WriteByte(0) // 1 byte reserved
	x.Body.Serialize(buf)

	data := buf.Bytes()

	length := uint16(len(data))
	putUint16(data, OSPFv3MessagePacketLengthAtByte, length)

	checksum := OSPFv3Checksum(data, src, dst)
	putUint16(data, OSPFv3MessageChecksumAtByte, checksum)

	out.Write(data)
}

func putUint16(b []byte, p int, v uint16) {
	binary.BigEndian.PutUint16(b[p:p+2], v)
}

func DeserializeOSPFv3Message(buf *bytes.Buffer, src, dst net.IP) (*OSPFv3Message, int, error) {
	pdu := &OSPFv3Message{}
	data := buf.Bytes()

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

	if pdu.Version != expectedVersion {
		return nil, readBytes, fmt.Errorf("Invalid OSPF version: %d", pdu.Version)
	}

	expectedChecksum := OSPFv3Checksum(data, src, dst)
	if pdu.Checksum != expectedChecksum {
		return nil, readBytes, fmt.Errorf("Checksum mismatch. Expected %#04x, got %#04x", expectedChecksum, pdu.Checksum)
	}

	n, err := pdu.ReadBody(buf)
	if err != nil {
		return nil, readBytes, errors.Wrap(err, "unable to decode message body")
	}
	readBytes += n

	return pdu, readBytes, nil
}

func OSPFv3Checksum(data []byte, src, dst net.IP) uint16 {
	data[12] = 0
	data[13] = 0
	return checksum.IPv6UpperLayerChecksum(src, dst, OSPFProtocolNumber, data)
}

func (m *OSPFv3Message) ReadBody(buf *bytes.Buffer) (int, error) {
	bodyLength := m.PacketLength - OSPFv3MessageHeaderLength
	var body Serializable
	var readBytes int
	var err error

	switch m.Type {
	case MsgTypeHello:
		body, readBytes, err = DeserializeHello(buf, bodyLength)
	case MsgTypeDatabaseDescription:
		body, readBytes, err = DeserializeDatabaseDescription(buf, bodyLength)
	case MsgTypeLinkStateRequest:
		body, readBytes, err = DeserializeLinkStateRequestMsg(buf, bodyLength)
	case MsgTypeLinkStateUpdate:
		body, readBytes, err = DeserializeLinkStateUpdate(buf)
	case MsgTypeLinkStateAcknowledgment:
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

type DBFlags uint8

// database description flags
const (
	DBFlagInit DBFlags = 1 << iota
	DBFlagMore
	DBFlagMS
)

type DatabaseDescription struct {
	Options          RouterOptions
	InterfaceMTU     uint16
	DBFlags          DBFlags
	DDSequenceNumber uint32
	LSAHeaders       []*LSA
}

func (x *DatabaseDescription) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(0) // 1 byte reserved
	x.Options.Serialize(buf)
	buf.Write(convert.Uint16Byte(x.InterfaceMTU))
	buf.WriteByte(0) // 1 byte reserved
	buf.WriteByte(uint8(x.DBFlags))
	buf.Write(convert.Uint32Byte(x.DDSequenceNumber))
	for i := range x.LSAHeaders {
		x.LSAHeaders[i].SerializeHeader(buf)
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
		tlv, n, err := DeserializeLSAHeader(buf)
		if err != nil {
			return nil, 0, errors.Wrap(err, "Unable to decode")
		}
		pdu.LSAHeaders = append(pdu.LSAHeaders, tlv)
		i += n
		readBytes += n
	}

	return pdu, readBytes, nil
}

type LinkStateRequestMsg []LinkStateRequest

func (x LinkStateRequestMsg) Serialize(buf *bytes.Buffer) {
	for i := range x {
		x[i].Serialize(buf)
	}
}

func DeserializeLinkStateRequestMsg(buf *bytes.Buffer, bodyLength uint16) (LinkStateRequestMsg, int, error) {
	reqs := make(LinkStateRequestMsg, 0)

	var readBytes int
	for readBytes < int(bodyLength) {
		req, n, err := DeserializeLinkStateRequest(buf)
		if err != nil {
			return nil, readBytes, errors.Wrap(err, "unable to decode LinkStateRequest")
		}
		reqs = append(reqs, req)
		readBytes += n
	}

	return reqs, readBytes, nil
}

type LinkStateRequest struct {
	LSType            LSAType
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

type LinkStateUpdate []*LSA

func (x LinkStateUpdate) Serialize(buf *bytes.Buffer) {
	buf.Write(convert.Uint32Byte(uint32(len(x))))
	for i := range x {
		x[i].Serialize(buf)
	}
}

func DeserializeLinkStateUpdate(buf *bytes.Buffer) (LinkStateUpdate, int, error) {
	lsas := make(LinkStateUpdate, 0)

	var lsaCount uint32
	if err := binary.Read(buf, binary.BigEndian, &lsaCount); err != nil {
		return nil, 0, errors.Wrap(err, "unable to decode LSA count")
	}
	readBytes := 4

	for i := 0; i < int(lsaCount); i++ {
		tlv, n, err := DeserializeLSA(buf)
		if err != nil {
			return nil, 0, errors.Wrap(err, "unable to decode LSA")
		}
		lsas = append(lsas, tlv)
		readBytes += n
	}

	return lsas, readBytes, nil
}

type LinkStateAcknowledgement struct {
	LSAHeaders []*LSA
}

func (x *LinkStateAcknowledgement) Serialize(buf *bytes.Buffer) {
	for i := range x.LSAHeaders {
		x.LSAHeaders[i].SerializeHeader(buf)
	}
}

func DeserializeLinkStateAcknowledgement(buf *bytes.Buffer, bodyLength uint16) (*LinkStateAcknowledgement, int, error) {
	pdu := &LinkStateAcknowledgement{}

	var readBytes int

	for i := 0; i < int(bodyLength); {
		tlv, n, err := DeserializeLSAHeader(buf)
		if err != nil {
			return nil, 0, errors.Wrap(err, "Unable to decode")
		}
		pdu.LSAHeaders = append(pdu.LSAHeaders, tlv)
		i += n
		readBytes += n
	}

	return pdu, readBytes, nil
}
