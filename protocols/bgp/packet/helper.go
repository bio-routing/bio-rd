package packet

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
)

func serializePrefix(pfx bnet.Prefix) []byte {
	if pfx.Pfxlen() == 0 {
		return []byte{}
	}

	numBytes := BytesInAddr(pfx.Pfxlen())

	b := make([]byte, numBytes+1)
	b[0] = pfx.Pfxlen()
	copy(b[1:numBytes+1], pfx.Addr().Bytes()[0:numBytes])

	return b
}

func deserializePrefix(b []byte, pfxLen uint8, afi uint16) (*bnet.Prefix, error) {
	numBytes := BytesInAddr(pfxLen)

	if numBytes != uint8(len(b)) {
		return nil, fmt.Errorf("could not parse prefix of length %d. Expected %d bytes, got %d", pfxLen, numBytes, len(b))
	}

	if afi == AFIIPv4 {
		return bnet.NewPfx(bnet.IPv4FromBytes(b), pfxLen).Dedup(), nil
	}

	ipBytes := make([]byte, afiAddrLenBytes[afi])
	copy(ipBytes, b)

	ip, err := bnet.IPFromBytes(ipBytes)
	if err != nil {
		return nil, err
	}

	pfx := bnet.NewPfx(ip, pfxLen)
	if !pfx.Valid() {
		return nil, fmt.Errorf("Invalid prefix: %q", pfx.String())
	}

	return pfx.Dedup(), nil
}

// REMOVE
