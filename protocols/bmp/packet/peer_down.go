package packet

import (
	"bytes"

	"github.com/bio-routing/bio-rd/util/decoder"
)

// PeerDownNotification represents a peer down notification
type PeerDownNotification struct {
	CommonHeader *CommonHeader
	Reason       uint8
	Data         []byte
}

// MsgType returns the type of this message
func (p *PeerDownNotification) MsgType() uint8 {
	return p.CommonHeader.MsgType
}

func decodePeerDownNotification(buf *bytes.Buffer, ch *CommonHeader) (*PeerDownNotification, error) {
	p := &PeerDownNotification{}

	fields := []interface{}{
		&p.Reason,
	}

	err := decoder.Decode(buf, fields)
	if err != nil {
		return nil, err
	}

	if p.Reason < 1 || p.Reason > 3 {
		return p, nil
	}

	p.Data = make([]byte, ch.MsgLength-CommonHeaderLen-1)
	fields = []interface{}{
		p.Data,
	}

	err = decoder.Decode(buf, fields)
	if err != nil {
		return nil, err
	}

	return p, nil
}