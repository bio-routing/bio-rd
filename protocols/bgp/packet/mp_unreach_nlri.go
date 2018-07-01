package packet

import (
	"bytes"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/taktv6/tflow2/convert"
)

// MultiProtocolUnreachNLRI represents network layer withdraw information for one prefix of an IP address family (rfc4760)
type MultiProtocolUnreachNLRI struct {
	AFI    uint16
	SAFI   uint8
	Prefix bnet.Prefix
}

func (n *MultiProtocolUnreachNLRI) serialize(buf *bytes.Buffer) uint8 {
	tempBuf := bytes.NewBuffer(nil)
	tempBuf.Write(convert.Uint16Byte(n.AFI))
	tempBuf.WriteByte(n.SAFI)
	tempBuf.Write(serializePrefix(n.Prefix))

	buf.Write(tempBuf.Bytes())

	return uint8(tempBuf.Len())
}
