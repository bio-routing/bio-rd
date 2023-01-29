package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"

	bnet "github.com/bio-routing/bio-rd/net"
)

// IPInterfaceAddressesTLVType is the type value of an IP interface address TLV
const IPInterfaceAddressesTLVType = 132

// IPInterfaceAddressesTLV represents an IP interface TLV
type IPInterfaceAddressesTLV struct {
	TLVType       uint8
	TLVLength     uint8
	IPv4Addresses []uint32
}

func NewIPInterfaceAddressesTLV(addrs []*bnet.Prefix) *IPInterfaceAddressesTLV {
	return &IPInterfaceAddressesTLV{
		TLVType:       IPInterfaceAddressesTLVType,
		TLVLength:     uint8(len(addrs) * 4),
		IPv4Addresses: ipv4AddrsFromPrefixes(addrs),
	}
}

func ipv4AddrsFromPrefixes(pfxs []*bnet.Prefix) []uint32 {
	addrs := make([]uint32, 0)
	for _, pfx := range pfxs {
		addrs = append(addrs, pfx.Addr().ToUint32())
	}

	return addrs
}

func readIPInterfaceAddressesTLV(buf *bytes.Buffer, tlvType uint8, tlvLength uint8) (*IPInterfaceAddressesTLV, error) {
	pdu := &IPInterfaceAddressesTLV{
		TLVType:       tlvType,
		TLVLength:     tlvLength,
		IPv4Addresses: make([]uint32, tlvLength/4),
	}

	fields := make([]interface{}, len(pdu.IPv4Addresses))
	for i := range pdu.IPv4Addresses {
		fields[i] = &pdu.IPv4Addresses[i]
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("unable to decode fields: %v", err)
	}

	return pdu, nil
}

func (i *IPInterfaceAddressesTLV) Copy() TLV {
	ret := *i
	ret.IPv4Addresses = make([]uint32, len(i.IPv4Addresses))
	copy(ret.IPv4Addresses, i.IPv4Addresses)
	return &ret
}

// Type returns the type of the TLV
func (i IPInterfaceAddressesTLV) Type() uint8 {
	return i.TLVType
}

// Length returns the length of the TLV
func (i IPInterfaceAddressesTLV) Length() uint8 {
	return i.TLVLength
}

// Value gets the TLV itself
func (i IPInterfaceAddressesTLV) Value() interface{} {
	return i
}

// Serialize serializes an IP interfaces address TLV
func (i IPInterfaceAddressesTLV) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(i.TLVType)
	buf.WriteByte(i.TLVLength)
	for j := range i.IPv4Addresses {
		buf.Write(convert.Uint32Byte(i.IPv4Addresses[j]))
	}
}
