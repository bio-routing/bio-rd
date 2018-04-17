package packet

import (
	"bytes"
	"fmt"
	"math"
	"net"
)

func decodeNLRIs(buf *bytes.Buffer, length uint16) (*NLRI, error) {
	var ret *NLRI
	var eol *NLRI
	var nlri *NLRI
	var err error
	var consumed uint8
	p := uint16(0)

	for p < length {
		nlri, consumed, err = decodeNLRI(buf)
		if err != nil {
			return nil, fmt.Errorf("Unable to decode NLRI: %v", err)
		}
		p += uint16(consumed)

		if ret == nil {
			ret = nlri
			eol = nlri
			continue
		}

		eol.Next = nlri
		eol = nlri
	}

	return ret, nil
}

func decodeNLRI(buf *bytes.Buffer) (*NLRI, uint8, error) {
	var addr [4]byte
	nlri := &NLRI{}

	err := decode(buf, []interface{}{&nlri.Pfxlen})
	if err != nil {
		return nil, 0, err
	}

	toCopy := uint8(math.Ceil(float64(nlri.Pfxlen) / float64(OctetLen)))
	for i := uint8(0); i < net.IPv4len%OctetLen; i++ {
		if i < toCopy {
			err := decode(buf, []interface{}{&addr[i]})
			if err != nil {
				return nil, 0, err
			}
		} else {
			addr[i] = 0
		}
	}
	nlri.IP = addr
	return nlri, toCopy + 1, nil
}
