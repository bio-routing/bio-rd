package packet

import (
	"bytes"
	"fmt"
	"math"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/tflow2/convert"
	"github.com/pkg/errors"
)

const (
	PathIdentifierLen = 4
	BytesPerLabel     = 3
	BitsPerLabel      = BytesPerLabel * 8
)

// NLRI represents a Network Layer Reachability Information
type NLRI struct {
	PathIdentifier uint32
	LabelStack     []LabelStackEntry
	Prefix         *bnet.Prefix
	Next           *NLRI
}

func decodeNLRIs(buf *bytes.Buffer, length uint16, afi uint16, safi uint8, addPath bool) (*NLRI, error) {
	var ret *NLRI
	var eol *NLRI
	var nlri *NLRI
	var err error
	var consumed uint8
	p := uint16(0)

	for p < length {
		nlri, consumed, err = decodeNLRI(buf, afi, safi, addPath)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to decode NLRI")
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

func decodeNLRI(buf *bytes.Buffer, afi uint16, safi uint8, addPath bool) (*NLRI, uint8, error) {
	nlri := &NLRI{}

	consumed := uint8(0)

	if addPath {
		err := decode.Decode(buf, []interface{}{
			&nlri.PathIdentifier,
		})
		if err != nil {
			return nil, consumed, errors.Wrap(err, "Unable to decode path identifier")
		}

		consumed += PathIdentifierLen
	}

	pfxLen, err := buf.ReadByte()
	if err != nil {
		return nil, consumed, err
	}
	consumed++

	if safi == LabeledUnicastSAFI {
		nlri.LabelStack = make([]LabelStackEntry, 0, 1)
		for {
			label, err := decodeLabel(buf)
			if err != nil {
				return nil, consumed, errors.Wrap(err, "decode label failed")
			}

			consumed += BytesPerLabel
			pfxLen -= BitsPerLabel
			nlri.LabelStack = append(nlri.LabelStack, label)

			if label.isBottomOfStack() {
				break
			}
		}
	}

	numBytes := uint8(BytesInAddr(pfxLen))
	bytes := make([]byte, numBytes)

	r, err := buf.Read(bytes)
	consumed += uint8(r)
	if r < int(numBytes) {
		return nil, consumed, fmt.Errorf("expected %d bytes for NLRI, only %d remaining", numBytes, r)
	}

	pfx, err := deserializePrefix(bytes, pfxLen, afi)
	if err != nil {
		return nil, consumed, err
	}
	nlri.Prefix = pfx

	return nlri, consumed, nil
}

func (n *NLRI) serialize(buf *bytes.Buffer, addPath bool, withLabeledUnicast bool) uint8 {
	numBytes := uint8(0)

	if addPath {
		buf.Write(convert.Uint32Byte(n.PathIdentifier))
		numBytes += 4
	}

	pfxLen := n.Prefix.Pfxlen()
	if withLabeledUnicast {
		pfxLen += uint8(len(n.LabelStack) * BitsPerLabel)
	}

	buf.WriteByte(pfxLen)
	numBytes++

	if withLabeledUnicast {
		labelCount := len(n.LabelStack)
		for i, l := range n.LabelStack {
			l.serialize(buf, i == labelCount-1)
			numBytes += BytesPerLabel
		}
	}

	pfxNumBytes := BytesInAddr(n.Prefix.Pfxlen())
	buf.Write(n.Prefix.Addr().Bytes()[:pfxNumBytes])
	numBytes += pfxNumBytes

	return numBytes
}

// BytesInAddr gets the amount of bytes needed to encode an NLRI of prefix length pfxlen
func BytesInAddr(pfxlen uint8) uint8 {
	return uint8(math.Ceil(float64(pfxlen) / 8))
}
