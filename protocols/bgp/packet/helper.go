package packet

import (
	"fmt"
	"math"

	bnet "github.com/bio-routing/bio-rd/net"
)

func serializePrefix(pfx bnet.Prefix) []byte {
	if pfx.Pfxlen() == 0 {
		return []byte{}
	}

	numBytes := uint8(math.Ceil(float64(pfx.Pfxlen()) / 8))

	b := make([]byte, numBytes+1)
	b[0] = pfx.Pfxlen()
	copy(b[1:numBytes+1], pfx.Addr().Bytes()[0:numBytes])

	return b
}

func deserializePrefix(b []byte, pfxLen uint8, afi uint16) (bnet.Prefix, error) {
	numBytes := int(math.Ceil(float64(pfxLen) / 8))

	if numBytes != len(b) {
		return bnet.Prefix{}, fmt.Errorf("could not parse prefix of legth %d. Expected %d bytes, got %d", pfxLen, numBytes, len(b))
	}

	ipBytes := make([]byte, afiAddrLenBytes[afi])
	copy(ipBytes, b)

	ip, err := bnet.IPFromBytes(ipBytes)
	if err != nil {
		return bnet.Prefix{}, err
	}

	return bnet.NewPfx(ip, pfxLen), nil
}
