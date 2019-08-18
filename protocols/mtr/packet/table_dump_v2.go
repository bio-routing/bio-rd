package packet

import (
	"bytes"
	"fmt"
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/pkg/errors"
)

func messageForTableDumpV2SubType(r RecordSubType) (MTRMessage, error) {
	switch r {
	case PEER_INDEX_TABLE:
		return &PeerIndexTable{}, nil
	case RIB_IPV4_UNICAST:
		return &RIBIPv4Unicast{}, nil
	case RIB_IPV4_MULTICAST:
		return &RIBIPv4Multicast{}, nil
	case RIB_IPV6_UNICAST:
		return &RIBIPv6Unicast{}, nil
	case RIB_IPV6_MULTICAST:
		return &RIBIPv6Multicast{}, nil
	case RIB_GENERIC:
		return &RIBGeneric{}, nil
	default:
		return nil, fmt.Errorf("unknown subtype %d for TABLE_DUMP_V2", r)
	}
}

type PeerIndexTable struct {
	BGPID       uint32
	NameLength  uint16
	Name        string
	PeerCount   uint16
	PeerEntries []PeerIndexTableEntry
}

type PeerIndexTableEntry struct {
	ASSizeBytes   uint8
	IPSizeBytes   uint8
	BGPID         uint32
	PeerIPAddress *net.IP
	PeerAs        []byte
}

func (p *PeerIndexTableEntry) Decode(data *bytes.Buffer) error {
	var flags byte
	err := decode.Decode(data, []interface{}{&flags, &p.BGPID})
	if err != nil {
		return errors.Wrap(err, "failed to decode peer table entry constant length values")
	}
	if flags&1 == 1 { // 7th bit is set
		p.IPSizeBytes = 16
	} else {
		p.IPSizeBytes = 4
	}
	if flags&2 == 2 { // 6th bit is set
		p.ASSizeBytes = 2
	} else {
		p.ASSizeBytes = 4
	}
	ipBytes := make([]byte, p.IPSizeBytes)
	_, err = data.Read(ipBytes)
	if err != nil {
		return errors.Wrap(err, "failed to read IP bytes from input")
	}
	p.PeerIPAddress, err = net.IPFromBytes(ipBytes)
	if err != nil {
		return errors.Wrap(err, "failed to convert bytes to IP")
	}
	p.PeerAs = make([]byte, p.ASSizeBytes)
	_, err = data.Read(p.PeerAs)
	if err != nil {
		return errors.Wrap(err, "failed to read AS bytes from input")
	}
	return nil
}

func (p *PeerIndexTable) Decode(data *bytes.Buffer) error {
	err := decode.Decode(data, []interface{}{&p.BGPID, &p.NameLength})
	if err != nil {
		return errors.Wrap(err, "failed to decode BGPID and length of name from input")
	}
	name := make([]byte, p.NameLength)
	_, err = data.Read(name)
	if err != nil {
		return errors.Wrap(err, "failed to read name bytes")
	}
	p.Name = string(name)
	err = decode.Decode(data, []interface{}{&p.PeerCount})
	if err != nil {
		return errors.Wrap(err, "failed to decode peer count")
	}
	p.PeerEntries = make([]PeerIndexTableEntry, p.PeerCount)
	for i := range p.PeerEntries {
		err = p.PeerEntries[i].Decode(data)
		if err != nil {
			return errors.Wrapf(err, "failed to decode %d peer table entry", i)
		}
	}
	return nil
}

type RIBEntry struct {
}

type AFI uint16

const (
	IPv4 AFI = 1
	IPv6 AFI = 2
)

// specificRIBSubType is a general implementation of the specific subtype
type specificRIBSubtype struct {
	SequenceNumber uint32
	Prefix         *net.Prefix
	RIBEntries     []RIBEntry
}

func (s *specificRIBSubtype) decodeSpecific(data *bytes.Buffer, afi AFI) error {
	err := decode.Decode(data, []interface{}{
		&s.SequenceNumber,
	})
	if err != nil {
		return errors.Wrap(err, "failed to decode sequence number")
	}
	s.Prefix, _, err = packet.DecodePrefixFromNLRI(data, uint16(afi))
	if err != nil {
		return errors.Wrap(err, "failed to decode prefix")
	}
	// TODO: Implement RIB entries
	return nil
}

type RIBIPv4Unicast struct {
	specificRIBSubtype
}

func (R *RIBIPv4Unicast) Decode(data *bytes.Buffer) error {
	return R.decodeSpecific(data, IPv4)
}

type RIBIPv4Multicast struct {
}

func (R *RIBIPv4Multicast) Decode(data *bytes.Buffer) error {
	return errors.New("Not implemented")
}

type RIBIPv6Unicast struct {
	specificRIBSubtype
}

func (R *RIBIPv6Unicast) Decode(data *bytes.Buffer) error {
	return R.decodeSpecific(data, IPv6)
}

type RIBIPv6Multicast struct {
}

func (R *RIBIPv6Multicast) Decode(data *bytes.Buffer) error {
	return errors.New("Not implemented")
}

type RIBGeneric struct {
}

func (R *RIBGeneric) Decode(data *bytes.Buffer) error {
	return errors.New("Not implemented")
}
