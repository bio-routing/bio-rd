package packet

import (
	"bytes"
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/taktv6/tflow2/convert"
)

// MultiProtocolUnreachNLRI represents network layer withdraw information for one prefix of an IP address family (rfc4760)
type MultiProtocolUnreachNLRI struct {
	AFI      uint16
	SAFI     uint8
	Prefixes []bnet.Prefix
	PathID   uint32
}

func (n *MultiProtocolUnreachNLRI) serialize(buf *bytes.Buffer, opt *EncodeOptions) uint16 {
	tempBuf := bytes.NewBuffer(nil)
	tempBuf.Write(convert.Uint16Byte(n.AFI))
	tempBuf.WriteByte(n.SAFI)
	for _, pfx := range n.Prefixes {
		if opt.UseAddPath {
			tempBuf.Write(convert.Uint32Byte(n.PathID))
		}
		tempBuf.Write(serializePrefix(pfx))
	}

	buf.Write(tempBuf.Bytes())

	return uint16(tempBuf.Len())
}

func deserializeMultiProtocolUnreachNLRI(b []byte) (MultiProtocolUnreachNLRI, error) {
	n := MultiProtocolUnreachNLRI{}

	prefixesLength := len(b) - 3 // 3 <- AFI + SAFI
	if prefixesLength < 0 {
		return n, fmt.Errorf("Invalid length of MP_UNREACH_NLRI: expected more than 3 bytes but got %d", len(b))
	}

	prefixes := make([]byte, prefixesLength)
	fields := []interface{}{
		&n.AFI,
		&n.SAFI,
		&prefixes,
	}
	err := decode(bytes.NewBuffer(b), fields)
	if err != nil {
		return MultiProtocolUnreachNLRI{}, err
	}

	if len(prefixes) == 0 {
		return n, nil
	}

	idx := uint16(0)
	for idx < uint16(len(prefixes)) {
		pfxLen := prefixes[idx]
		numBytes := uint16(BytesInAddr(pfxLen))
		idx++

		r := uint16(len(prefixes)) - idx
		if r < numBytes {
			return MultiProtocolUnreachNLRI{}, fmt.Errorf("expected %d bytes for NLRI, only %d remaining", numBytes, r)
		}

		start := idx
		end := idx + numBytes
		pfx, err := deserializePrefix(prefixes[start:end], pfxLen, n.AFI)
		if err != nil {
			return MultiProtocolUnreachNLRI{}, err
		}
		n.Prefixes = append(n.Prefixes, pfx)

		idx = idx + numBytes
	}

	return n, nil
}
