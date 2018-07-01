package packet

import (
	"math"

	bnet "github.com/bio-routing/bio-rd/net"
)

func serializePrefix(pfx bnet.Prefix) []byte {
	if pfx.Pfxlen() == 0 {
		return []byte{}
	}

	numBytes := uint8(math.Ceil(float64(pfx.Pfxlen()) / float64(8)))

	b := make([]byte, numBytes+1)
	b[0] = pfx.Pfxlen()
	copy(b[1:numBytes+1], pfx.Addr().Bytes()[0:numBytes])

	return b
}
