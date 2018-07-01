package packet

import (
	"bytes"
	"fmt"

	"github.com/taktv6/tflow2/convert"

	bnet "github.com/bio-routing/bio-rd/net"
)

// MultiProtocolReachNLRI represents network layer reachability information for one prefix of an IP address family (rfc4760)
type MultiProtocolReachNLRI struct {
	AFI     uint16
	SAFI    uint8
	NextHop bnet.IP
	Prefix  bnet.Prefix
}

func (n *MultiProtocolReachNLRI) serialize(buf *bytes.Buffer) uint8 {
	nextHop := n.NextHop.Bytes()

	tempBuf := bytes.NewBuffer(nil)
	tempBuf.Write(convert.Uint16Byte(n.AFI))
	tempBuf.WriteByte(n.SAFI)
	tempBuf.WriteByte(uint8(len(nextHop))) // NextHop length
	tempBuf.Write(nextHop)
	tempBuf.WriteByte(0) // RESERVED
	tempBuf.Write(serializePrefix(n.Prefix))

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

	n.Prefix, err = deserializePrefix(variable[nextHopLength+2:], variable[nextHopLength+1], n.AFI)
	if err != nil {
		return MultiProtocolReachNLRI{}, err
	}

	return n, nil
}
