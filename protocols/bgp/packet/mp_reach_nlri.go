package packet

import (
	"bytes"
	"math"

	"github.com/taktv6/tflow2/convert"

	bnet "github.com/bio-routing/bio-rd/net"
)

// MultiProtocolReachNLRI represents Network Layer Reachability Information for one prefix of an IP address family (rfc4760)
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
	tempBuf.Write(n.serializePrefix())

	buf.Write(tempBuf.Bytes())

	return uint8(tempBuf.Len())
}

func (n *MultiProtocolReachNLRI) serializePrefix() []byte {
	if n.Prefix.Pfxlen() == 0 {
		return []byte{}
	}

	numBytes := uint8(math.Ceil(float64(n.Prefix.Pfxlen()) / float64(8)))

	b := make([]byte, numBytes+1)
	b[0] = n.Prefix.Pfxlen()
	copy(b[1:numBytes+1], n.Prefix.Addr().Bytes()[0:numBytes])

	return b
}
