package packet

import (
	"bytes"
	"fmt"
	"time"
)

type RecordType uint16

const (
	OSPFv2        RecordType = 11
	TABLE_DUMP    RecordType = 12
	TABLE_DUMP_V2 RecordType = 13
	BGP4MP        RecordType = 16
	BGP4MP_ET     RecordType = 17
	ISIS          RecordType = 32
	ISIS_ET       RecordType = 33
	OSPFv3        RecordType = 48
	OSPFv3_ET     RecordType = 48
)

type RecordSubType uint16

const (
	PEER_INDEX_TABLE   RecordSubType = 1
	RIB_IPV4_UNICAST   RecordSubType = 2
	RIB_IPV4_MULTICAST RecordSubType = 3
	RIB_IPV6_UNICAST   RecordSubType = 4
	RIB_IPV6_MULTICAST RecordSubType = 5
	RIB_GENERIC        RecordSubType = 6
)

type MessageType struct {
	Type    RecordType
	SubType RecordSubType
}

// An MTRMessage decodes itself from the given Reader. It will not read until
// the end of the reader. Instead, it reads only the parts defined by the MTR header
type MTRMessage interface {
	Decode(data *bytes.Buffer) error
}

type MTRRecord struct {
	TimeStamp time.Time
	Type      MessageType
	Length    uint32
	Message   MTRMessage
}

func messageForType(t MessageType) (MTRMessage, error) {
	switch t.Type {
	case TABLE_DUMP_V2:
		return messageForTableDumpV2SubType(t.SubType)
	default:
		return nil, fmt.Errorf("given type %d subtype %d not implemented", t.Type, t.SubType)
	}
}
