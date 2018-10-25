package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decoder"
)

const (
	reasonMin = 1
	reasonMax = 3
)

// PeerDownNotification represents a peer down notification
type PeerDownNotification struct {
	CommonHeader  *CommonHeader
	PerPeerHeader *PerPeerHeader
	Reason        uint8
	Data          []byte
}

// MsgType returns the type of this message
func (p *PeerDownNotification) MsgType() uint8 {
	return p.CommonHeader.MsgType
}

func decodePeerDownNotification(buf *bytes.Buffer, ch *CommonHeader) (*PeerDownNotification, error) {
	p := &PeerDownNotification{
		CommonHeader: ch,
	}

	pph, err := decodePerPeerHeader(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode per peer header: %v", err)
	}

	p.PerPeerHeader = pph

	fields := []interface{}{
		&p.Reason,
	}

	err = decoder.Decode(buf, fields)
	if err != nil {
		return nil, err
	}

	if p.Reason < reasonMin || p.Reason > reasonMax {
		return p, nil
	}

	p.Data = make([]byte, ch.MsgLength-PerPeerHeaderLen-CommonHeaderLen-1)
	fields = []interface{}{
		p.Data,
	}

	err = decoder.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to read Data: %v", err)
	}

	return p, nil
}
