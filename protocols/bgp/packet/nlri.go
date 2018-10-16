package packet

import (
	"bytes"
	"fmt"
	"math"
	"net"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/taktv6/tflow2/convert"
)

type NLRI struct {
	PathIdentifier uint32
	IP             bnet.IP
	Pfxlen         uint8
	Next           *NLRI
}

func decodeNLRIs(buf *bytes.Buffer, length uint16) (*NLRI, error) {
	var ret *NLRI
	var eol *NLRI
	var nlri *NLRI
	var err error
	var consumed uint8
	p := uint16(0)

	for p < length {
		nlri, consumed, err = decodeNLRI(buf)
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

func decodeNLRI(buf *bytes.Buffer) (*NLRI, uint8, error) {
	addr := make([]byte, 4)
	nlri := &NLRI{}

	err := decode.Decode(buf, []interface{}{&nlri.Pfxlen})
	if err != nil {
		return nil, 0, err
	}

	toCopy := uint8(math.Ceil(float64(nlri.Pfxlen) / float64(OctetLen)))
	for i := uint8(0); i < net.IPv4len%OctetLen; i++ {
		if i < toCopy {
			err := decode.Decode(buf, []interface{}{&addr[i]})
			if err != nil {
				return nil, 0, err
			}
		} else {
			addr[i] = 0
		}
	}
	nlri.IP, err = bnet.IPFromBytes(addr)
	if err != nil {
		return nil, 0, err
	}

	return nlri, toCopy + 1, nil
}

func (n *NLRI) serialize(buf *bytes.Buffer) uint8 {
	buf.WriteByte(n.Pfxlen)
	b := n.IP.Bytes()

	nBytes := BytesInAddr(n.Pfxlen)
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
