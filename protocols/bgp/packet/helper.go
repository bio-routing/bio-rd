package packet

import (
	"fmt"

	bnet "github.com/bio-routing/bio-rd/net"
)

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
