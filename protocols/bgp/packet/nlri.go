package packet

import (
	"bytes"
	"fmt"
	"math"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/taktv6/tflow2/convert"
)

type NLRI struct {
	PathIdentifier uint32
	Prefix         bnet.Prefix
	Next           *NLRI
}

func decodeNLRIs(buf *bytes.Buffer, length uint16, afi uint16) (*NLRI, error) {
	var ret *NLRI
	var eol *NLRI
	var nlri *NLRI
	var err error
	var consumed uint8
	p := uint16(0)

	for p < length {
		nlri, consumed, err = decodeNLRI(buf, afi)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode NLRI: %v", err)
		}
		p += uint16(consumed)

		if ret == nil {
			ret = nlri
			eol = nlri
			continue
		}

		eol.Next = nlri
		eol = nlri
	}

	return ret, nil
}

func decodeNLRI(buf *bytes.Buffer, afi uint16) (*NLRI, uint8, error) {
	nlri := &NLRI{}

	consumed := uint8(0)
	pfxLen, err := buf.ReadByte()
	if err != nil {
		return nil, consumed, err
	}
	consumed++

	numBytes := uint8(BytesInAddr(pfxLen))
	bytes := make([]byte, numBytes)

	r, err := buf.Read(bytes)
	consumed += uint8(r)
	if r < int(numBytes) {
		return nil, consumed, fmt.Errorf("expected %d bytes for NLRI, only %d remaining", numBytes, r)
	}

	pfx, err := deserializePrefix(bytes, pfxLen, afi)
	if err != nil {
		return nil, consumed, err
	}
	nlri.Prefix = pfx

	return nlri, consumed, nil
}

func (n *NLRI) serialize(buf *bytes.Buffer) uint8 {
	buf.WriteByte(n.Prefix.Pfxlen())
	b := n.Prefix.Addr().Bytes()

	nBytes := BytesInAddr(n.Prefix.Pfxlen())
	buf.Write(b[:nBytes])

	return nBytes + 1
}

func (n *NLRI) serializeAddPath(buf *bytes.Buffer) uint8 {
	buf.Write(convert.Uint32Byte(n.PathIdentifier))

	return uint8(n.serialize(buf) + 4)
}

// BytesInAddr gets the amount of bytes needed to encode an NLRI of prefix length pfxlen
func BytesInAddr(pfxlen uint8) uint8 {
	return uint8(math.Ceil(float64(pfxlen) / 8))
}
