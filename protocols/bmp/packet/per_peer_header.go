package packet

import (
	"bytes"

	"github.com/bio-routing/bio-rd/util/decoder"
	"github.com/taktv6/tflow2/convert"
)

const (
	// PerPeerHeaderLen is the length of a per peer header
	PerPeerHeaderLen = 42
)

// PerPeerHeader represents a BMP per peer header
type PerPeerHeader struct {
	PeerType              uint8
	PeerFlags             uint8
	PeerDistinguisher     uint64
	PeerAddress           [16]byte
	PeerAS                uint32
	PeerBGPID             uint32
	Timestamp             uint32
	TimestampMicroSeconds uint32
}

// Serialize serializes a per peer header
func (p *PerPeerHeader) Serialize(buf *bytes.Buffer) {
	buf.WriteByte(p.PeerType)
	buf.WriteByte(p.PeerFlags)
	buf.Write(convert.Uint64Byte(p.PeerDistinguisher))
	buf.Write(p.PeerAddress[:])
	buf.Write(convert.Uint32Byte(p.PeerAS))
	buf.Write(convert.Uint32Byte(p.PeerBGPID))
	buf.Write(convert.Uint32Byte(p.Timestamp))
	buf.Write(convert.Uint32Byte(p.TimestampMicroSeconds))
}

func decodePerPeerHeader(buf *bytes.Buffer) (*PerPeerHeader, error) {
	p := &PerPeerHeader{}

	fields := []interface{}{
		&p.PeerType,
		&p.PeerFlags,
		&p.PeerDistinguisher,
		&p.PeerAddress,
		&p.PeerAS,
		&p.PeerBGPID,
		&p.Timestamp,
		&p.TimestampMicroSeconds,
	}

	err := decoder.Decode(buf, fields)
	if err != nil {
		return p, err
	}

	return p, nil
}

// GetIPVersion gets the IP version of the BGP session
func (p *PerPeerHeader) GetIPVersion() uint8 {
	if p.PeerFlags>>7 == 1 {
		return 6
	}
	return 4
}
