package packetv3

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"
)

// Prefix Option Bits
const (
	NUBIT = 1
	LABIT = 2
	PBIT  = 8
	DNBIT = 16
)

type PrefixOptions struct {
	NoUnicast    bool // NU-bit
	LocalAddress bool // LA-bit
	Propagate    bool // P-bit
	DN           bool // DN-bit
}

func (o PrefixOptions) Serialize(buf *bytes.Buffer) {
	var rawOptions uint8
	if o.NoUnicast {
		rawOptions = rawOptions | NUBIT
	}
	if o.LocalAddress {
		rawOptions = rawOptions | LABIT
	}
	if o.Propagate {
		rawOptions = rawOptions | PBIT
	}
	if o.DN {
		rawOptions = rawOptions | DNBIT
	}
	buf.WriteByte(rawOptions)
}

type LSAPrefix struct {
	PrefixLength uint8
	Options      PrefixOptions

	// this may represent different things
	// used for metric or referenced LSType
	Special uint16

	Address net.IP
}

func DeserializeLSAPrefix(buf *bytes.Buffer) (LSAPrefix, int, error) {
	pdu := LSAPrefix{}

	var readBytes int
	var err error

	var rawOptions uint8

	fields := []interface{}{
		&pdu.PrefixLength,
		&rawOptions,
		&pdu.Special,
	}

	err = decode.Decode(buf, fields)
	if err != nil {
		return pdu, readBytes, fmt.Errorf("Unable to decode fields: %v", err)
	}
	readBytes += 4

	// read Options
	pdu.Options.NoUnicast = (rawOptions & NUBIT) != 0
	pdu.Options.LocalAddress = (rawOptions & LABIT) != 0
	pdu.Options.Propagate = (rawOptions & PBIT) != 0
	pdu.Options.DN = (rawOptions & DNBIT) != 0

	// read AddressPrefix
	numBytes := int((pdu.PrefixLength+31)/32) * 4
	pfxBytes := buf.Next(numBytes)
	ipBytes := make([]byte, 16)
	copy(ipBytes[:len(pfxBytes)], pfxBytes)
	addr, err := net.IPFromBytes(ipBytes)
	if err != nil {
		return pdu, readBytes, errors.Wrap(err, "unable to decode AddressPrefix")
	}
	pdu.Address = *addr
	readBytes += len(pfxBytes)

	return pdu, readBytes, nil
}

func (x *LSAPrefix) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(x.PrefixLength)
	x.Options.Serialize(buf)
	buf.Write(convert.Uint16Byte(x.Special))

	// serialize AddressPrefix
	numBytes := int((x.PrefixLength+31)/32) * 4
	buf.Write(x.Address.Bytes()[:numBytes])
}
