package packet

import (
	"bytes"
	"fmt"

	"github.com/taktv6/tflow2/convert"

	bnet "github.com/bio-routing/bio-rd/net"
)

// MultiProtocolReachNLRI represents network layer reachability information for one prefix of an IP address family (rfc4760)
type MultiProtocolReachNLRI struct {
	AFI      uint16
	SAFI     uint8
	NextHop  bnet.IP
	Prefixes []bnet.Prefix
}

func (n *MultiProtocolReachNLRI) serialize(buf *bytes.Buffer) uint8 {
	nextHop := n.NextHop.Bytes()

	tempBuf := bytes.NewBuffer(nil)
	tempBuf.Write(convert.Uint16Byte(n.AFI))
	tempBuf.WriteByte(n.SAFI)
	tempBuf.WriteByte(uint8(len(nextHop))) // NextHop length
	tempBuf.Write(nextHop)
	tempBuf.WriteByte(0) // RESERVED
	for _, pfx := range n.Prefixes {
		tempBuf.Write(serializePrefix(pfx))
	}

	buf.Write(tempBuf.Bytes())

	return uint8(tempBuf.Len())
}

func deserializeMultiProtocolReachNLRI(b []byte) (MultiProtocolReachNLRI, error) {
	n := MultiProtocolReachNLRI{}
	nextHopLength := uint8(0)
	variable := make([]byte, len(b)-4)

	fields := []interface{}{
		&n.AFI,
		&n.SAFI,
		&nextHopLength,
		&variable,
	}
	err := decode(bytes.NewBuffer(b), fields)
	if err != nil {
		return MultiProtocolReachNLRI{}, err
	}

	n.NextHop, err = bnet.IPFromBytes(variable[:nextHopLength])
	if err != nil {
		return MultiProtocolReachNLRI{}, fmt.Errorf("Failed to decode next hop IP: %v", err)
	}

	variable = variable[1+nextHopLength:]

	idx := uint8(0)
	n.Prefixes = make([]bnet.Prefix, 0)
	for idx < uint8(len(variable)) {
		l := numberOfBytesForPrefixLength(variable[idx])

		pfx, err := deserializePrefix(variable[idx+1:idx+1+l], variable[idx], n.AFI)
		if err != nil {
			return MultiProtocolReachNLRI{}, err
		}
		n.Prefixes = append(n.Prefixes, pfx)

		idx = idx + l + 1
	}

	return n, nil
}
