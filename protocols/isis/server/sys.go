package server

import (
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/device"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
)

type sys interface {
	openPacketSocket() error
	closePacketSocket() error
	mcastJoin(addr [6]byte) error
	sendPacket(pkt []byte, dst [6]byte) error
	recvPacket() (pkt []byte, src types.MACAddress, err error)
}

type bioSys struct {
	socket int
	device *device.Device
}

type mockSys struct {
	wantFailOpenPacketSocket   bool
	wantFailClosedPacketSocket bool
	wantFailMcastJoin          bool
	wantFailSendPacket         bool
	wantFailRecvPacket         bool
	recvPktCount               int
	recvPktPkt                 []byte
	stopRecvPktFail            chan struct{}
	stopRecvPktGraceful        chan struct{}
	closePacketSocketCalled    bool
}

func (m *mockSys) openPacketSocket() error {
	if m.wantFailOpenPacketSocket {
		return fmt.Errorf("Fail")
	}

	return nil
}

func (m *mockSys) closePacketSocket() error {
	m.closePacketSocketCalled = true
	if m.wantFailClosedPacketSocket {
		return fmt.Errorf("Fail")
	}

	return nil
}

func (m *mockSys) mcastJoin(addr [6]byte) error {
	if m.wantFailMcastJoin {
		return fmt.Errorf("Fail")
	}

	return nil
}

func (m *mockSys) sendPacket(pkt []byte, dst [6]byte) error {
	if m.wantFailSendPacket {
		return fmt.Errorf("Fail")
	}

	return nil
}

func (m *mockSys) recvPacket() (pkt []byte, src types.MACAddress, err error) {
	if m.wantFailRecvPacket {
		select {
		case <-m.stopRecvPktFail:
			return nil, [6]byte{}, fmt.Errorf("Fail")
		case <-m.stopRecvPktGraceful:
			return nil, [6]byte{}, fmt.Errorf("Stopped")
		}
	}

	if m.recvPktCount == 1 {
		select {
		case <-m.stopRecvPktFail:
			return nil, [6]byte{}, fmt.Errorf("Socket closed")
		case <-m.stopRecvPktGraceful:
			return m.recvPktPkt, [6]byte{10, 20, 30, 40, 50, 60}, nil
		}
	}

	m.recvPktCount--
	return m.recvPktPkt, [6]byte{10, 20, 30, 40, 50, 60}, nil
}
