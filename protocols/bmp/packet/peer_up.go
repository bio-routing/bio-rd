package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/util/decoder"
)

const (
	// OpenMsgMinLen is the minimal length of a BGP open message
	OpenMsgMinLen = 29
)

// PeerUpNotification represents a peer up notification
type PeerUpNotification struct {
	CommonHeader    *CommonHeader
	PerPeerHeader   *PerPeerHeader
	LocalAddress    [16]byte
	LocalPort       uint16
	RemotePort      uint16
	SentOpenMsg     []byte
	ReceivedOpenMsg []byte
	Information     []byte
}

// MsgType returns the type of this message
func (p *PeerUpNotification) MsgType() uint8 {
	return p.CommonHeader.MsgType
}

func decodePeerUpNotification(buf *bytes.Buffer, ch *CommonHeader) (*PeerUpNotification, error) {
	p := &PeerUpNotification{
		CommonHeader: ch,
	}

	pph, err := decodePerPeerHeader(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode per peer header: %v", err)
	}

	p.PerPeerHeader = pph

	fields := []interface{}{
		&p.LocalAddress,
		&p.LocalPort,
		&p.RemotePort,
	}

	err = decoder.Decode(buf, fields)
	if err != nil {
		return nil, err
	}

	sentOpenMsg, err := getOpenMsg(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to get OPEN message: %v", err)
	}
	p.SentOpenMsg = sentOpenMsg

	recvOpenMsg, err := getOpenMsg(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to get OPEN message: %v", err)
	}
	p.ReceivedOpenMsg = recvOpenMsg

	if buf.Len() == 0 {
		return p, nil
	}

	p.Information = make([]byte, buf.Len())
	fields = []interface{}{
		&p.Information,
	}

	// This can not fail as p.Information has exactly the size of what is left in buf
	decoder.Decode(buf, fields)

	return p, nil
}

func getOpenMsg(buf *bytes.Buffer) ([]byte, error) {
	msg := make([]byte, OpenMsgMinLen)

	fields := []interface{}{
		&msg,
	}
	err := decoder.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to read: %v", err)
	}

	if msg[OpenMsgMinLen-1] == 0 {
		return msg, nil
	}

	optParams := make([]byte, msg[OpenMsgMinLen-1])
	fields = []interface{}{
		&optParams,
	}

	err = decoder.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to read: %v", err)
	}

	msg = append(msg, optParams...)
	return msg, nil
}
