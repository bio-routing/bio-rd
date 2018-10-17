package packet

import (
	"bytes"
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/taktv6/tflow2/convert"
)

// MultiProtocolReachNLRI represents network layer reachability information for an IP address family (rfc4760)
type MultiProtocolReachNLRI struct {
	AFI     uint16
	SAFI    uint8
	NextHop bnet.IP
	NLRI    *NLRI
}

func (n *MultiProtocolReachNLRI) serialize(buf *bytes.Buffer, opt *EncodeOptions) uint16 {
	nextHop := n.NextHop.Bytes()

	tempBuf := bytes.NewBuffer(nil)
	tempBuf.Write(convert.Uint16Byte(n.AFI))
	tempBuf.WriteByte(n.SAFI)
	tempBuf.WriteByte(uint8(len(nextHop)))
	tempBuf.Write(nextHop)
	tempBuf.WriteByte(0) // RESERVED

	for cur := n.NLRI; cur != nil; cur = cur.Next {
		cur.serialize(tempBuf, opt.UseAddPath)
	}

	buf.Write(tempBuf.Bytes())

	return uint16(tempBuf.Len())
}

func deserializeMultiProtocolReachNLRI(b []byte, addPath bool) (MultiProtocolReachNLRI, error) {
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
	err := decode.Decode(bytes.NewBuffer(b), fields)
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

	if budget == 0 {
		return n, nil
	}

	variable = variable[1+nextHopLength:] // 1 <- RESERVED field

	buf := bytes.NewBuffer(variable)
	nlri, err := decodeNLRIs(buf, uint16(buf.Len()), n.AFI, addPath)
	if err != nil {
		return MultiProtocolReachNLRI{}, err
	}
	n.NLRI = nlri

	return n, nil
}
