package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"
)

// MultiProtocolUnreachNLRI represents network layer withdraw information for one prefix of an IP address family (rfc4760)
type MultiProtocolUnreachNLRI struct {
	AFI  uint16
	SAFI uint8
	NLRI *NLRI
}

func (n *MultiProtocolUnreachNLRI) serialize(buf *bytes.Buffer, opt *EncodeOptions) uint16 {
	tempBuf := bytes.NewBuffer(nil)
	tempBuf.Write(convert.Uint16Byte(n.AFI))
	tempBuf.WriteByte(n.SAFI)

	for cur := n.NLRI; cur != nil; cur = cur.Next {
		cur.serialize(tempBuf, opt.UseAddPath, n.SAFI)
	}

	buf.Write(tempBuf.Bytes())

	return uint16(tempBuf.Len())
}

func deserializeMultiProtocolUnreachNLRI(b []byte, opt *DecodeOptions) (MultiProtocolUnreachNLRI, error) {
	n := MultiProtocolUnreachNLRI{}

	prefixesLength := len(b) - 3 // 3 <- AFI + SAFI
	if prefixesLength < 0 {
		return n, fmt.Errorf("Invalid length of MP_UNREACH_NLRI: expected more than 3 bytes but got %d", len(b))
	}

	nlris := make([]byte, prefixesLength)
	fields := []interface{}{
		&n.AFI,
		&n.SAFI,
		&nlris,
	}
	err := decode.Decode(bytes.NewBuffer(b), fields)
	if err != nil {
		return MultiProtocolUnreachNLRI{}, err
	}

	if len(nlris) == 0 {
		return n, nil
	}

	buf := bytes.NewBuffer(nlris)
	nlri, err := decodeNLRIs(buf, uint16(buf.Len()), n.AFI, n.SAFI, opt.addPath(n.AFI, n.SAFI))
	if err != nil {
		return MultiProtocolUnreachNLRI{}, err
	}
	n.NLRI = nlri

	return n, nil
}
