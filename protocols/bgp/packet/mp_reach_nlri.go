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
	PathID   uint32
}

func (n *MultiProtocolReachNLRI) serialize(buf *bytes.Buffer, opt *EncodeOptions) uint16 {
	nextHop := n.NextHop.Bytes()

	tempBuf := bytes.NewBuffer(nil)
	tempBuf.Write(convert.Uint16Byte(n.AFI))
	tempBuf.WriteByte(n.SAFI)
	tempBuf.WriteByte(uint8(len(nextHop)))
	tempBuf.Write(nextHop)
	tempBuf.WriteByte(0) // RESERVED
	for _, pfx := range n.Prefixes {
		if opt.UseAddPath {
			tempBuf.Write(convert.Uint32Byte(n.PathID))
		}
		tempBuf.Write(serializePrefix(pfx))
	}

	buf.Write(tempBuf.Bytes())

	return uint16(tempBuf.Len())
}

func deserializeMultiProtocolReachNLRI(b []byte) (MultiProtocolReachNLRI, error) {
	n := MultiProtocolReachNLRI{}
	nextHopLength := uint8(0)

	variableLength := len(b) - 4 // 4 <- AFI + SAFI + NextHopLength
	if variableLength <= 0 {
		return n, fmt.Errorf("Invalid length of MP_REACH_NLRI: expected more than 4 bytes but got %d", len(b))
	}

	variable := make([]byte, variableLength)
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

	budget := variableLength
	if budget < int(nextHopLength) {
		return MultiProtocolReachNLRI{},
			fmt.Errorf("Failed to decode next hop IP: expected %d bytes for NLRI, only %d remaining", nextHopLength, budget)
	}

	n.NextHop, err = bnet.IPFromBytes(variable[:nextHopLength])
	if err != nil {
		return MultiProtocolReachNLRI{}, fmt.Errorf("Failed to decode next hop IP: %v", err)
	}
	budget -= int(nextHopLength)

	n.Prefixes = make([]bnet.Prefix, 0)
	if budget == 0 {
		return n, nil
	}

	variable = variable[1+nextHopLength:] // 1 <- RESERVED field

	idx := uint16(0)
	for idx < uint16(len(variable)) {
		pfxLen := variable[idx]
		numBytes := uint16(BytesInAddr(pfxLen))
		idx++

		r := uint16(len(variable)) - idx
		if r < numBytes {
			return MultiProtocolReachNLRI{}, fmt.Errorf("expected %d bytes for NLRI, only %d remaining", numBytes, r)
		}

		start := idx
		end := idx + numBytes
		pfx, err := deserializePrefix(variable[start:end], pfxLen, n.AFI)
		if err != nil {
			return MultiProtocolReachNLRI{}, err
		}
		n.Prefixes = append(n.Prefixes, pfx)

		idx = idx + numBytes
	}

	return n, nil
}
