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
}

func (n *MultiProtocolUnreachNLRI) serialize(buf *bytes.Buffer) uint16 {
	tempBuf := bytes.NewBuffer(nil)
	tempBuf.Write(convert.Uint16Byte(n.AFI))
	tempBuf.WriteByte(n.SAFI)
	for _, pfx := range n.Prefixes {
		tempBuf.Write(serializePrefix(pfx))
	}

	buf.Write(tempBuf.Bytes())

	return uint16(tempBuf.Len())
}

func deserializeMultiProtocolUnreachNLRI(b []byte) (MultiProtocolUnreachNLRI, error) {
	n := MultiProtocolUnreachNLRI{}
	prefix := make([]byte, len(b)-3)

	fields := []interface{}{
		&n.AFI,
		&n.SAFI,
		&prefix,
	}
	err := decode(bytes.NewBuffer(b), fields)
	if err != nil {
		return MultiProtocolUnreachNLRI{}, err
	}

	if len(prefix) == 0 {
		return n, nil
	}

	idx := uint16(0)
	for idx < uint16(len(prefix)) {
		pfxLen := prefix[idx]
		numBytes := uint16(BytesInAddr(pfxLen))
		idx++

		r := uint16(len(prefix)) - idx
		if r < numBytes {
			return MultiProtocolUnreachNLRI{}, fmt.Errorf("expected %d bytes for NLRI, only %d remaining", numBytes, r)
		}

		start := idx
		end := idx + numBytes
		pfx, err := deserializePrefix(prefix[start:end], pfxLen, n.AFI)
		if err != nil {
			return MultiProtocolUnreachNLRI{}, err
		}
		n.Prefixes = append(n.Prefixes, pfx)

		idx = idx + numBytes
	}

	return n, nil
}
