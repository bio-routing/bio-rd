package ethernet

import (
	"fmt"
)

const MockEthernetInterfaceBufferSize = 1024

type MockEthernetInterface struct {
	joinedMCastGroups []MACAddr
	sendCh            chan mockPkt
	recvCh            chan mockPkt
	closedCh          chan struct{}
}

type mockPkt struct {
	mac    MACAddr
	packet []byte
}

func NewMockEthernetInterface() *MockEthernetInterface {
	return &MockEthernetInterface{
		joinedMCastGroups: make([]MACAddr, 0),
		sendCh:            make(chan mockPkt, MockEthernetInterfaceBufferSize),
		recvCh:            make(chan mockPkt, MockEthernetInterfaceBufferSize),
		closedCh:          make(chan struct{}),
	}
}

func (mei *MockEthernetInterface) RecvPacket() (pkt []byte, src MACAddr, err error) {
	select {
	case <-mei.closedCh:
		return nil, MACAddr{}, fmt.Errorf("socket closed")
	case p := <-mei.recvCh:
		return p.packet, p.mac, nil
	}
}

func (mei *MockEthernetInterface) SendPacket(dst MACAddr, pkt []byte) error {
	select {
	case <-mei.closedCh:
		return fmt.Errorf("socket closed")
	case mei.sendCh <- mockPkt{
		mac:    dst,
		packet: pkt,
	}:
		return nil
	}
}

func (mei *MockEthernetInterface) MCastJoin(addr MACAddr) error {
	mei.joinedMCastGroups = append(mei.joinedMCastGroups, addr)

	return nil
}

func (mei *MockEthernetInterface) GetMTU() int {
	return 1500
}

func (mei *MockEthernetInterface) Close() {
	close(mei.closedCh)
}

func (mei *MockEthernetInterface) SendFromRemote(src MACAddr, pkt []byte) {
	mei.recvCh <- mockPkt{
		mac:    src,
		packet: pkt,
	}
}

func (mei *MockEthernetInterface) ReceiveAtRemote() (MACAddr, []byte) {
	p := <-mei.sendCh
	return p.mac, p.packet
}

func (mei *MockEthernetInterface) DrainBuffer() {
	for {
		select {
		case <-mei.sendCh:
			continue
		default:
			return
		}
	}
}
