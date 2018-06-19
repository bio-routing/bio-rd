package server

import (
	"fmt"
	"net"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
)

type state interface {
	run() (state, string)
}

// FSM2 implements the BGP finite state machine (RFC4271)
type FSM2 struct {
	server  *BGPServer
	peer    *Peer
	eventCh chan int
	con     net.Conn
	conCh   chan net.Conn

	delayOpen      bool
	delayOpenTime  time.Duration
	delayOpenTimer *time.Timer

	connectRetryTime    time.Duration
	connectRetryTimer   *time.Timer
	connectRetryCounter int

	holdTimeConfigured time.Duration
	holdTime           time.Duration
	holdTimer          *time.Timer

	keepaliveTime  time.Duration
	keepaliveTimer *time.Timer

	msgRecvCh     chan msgRecvMsg
	msgRecvFailCh chan msgRecvErr
	stopMsgRecvCh chan struct{}

	capAddPathSend bool
	capAddPathRecv bool

	local  net.IP
	remote net.IP

	ribsInitialized bool
	adjRIBIn        *adjRIBIn.AdjRIBIn
	adjRIBOut       routingtable.RouteTableClient
	rib             *locRIB.LocRIB
	updateSender    routingtable.RouteTableClient

	neighborID uint32
	state      state
	reason     string
	active     bool
}

// NewPassiveFSM2 initiates a new passive FSM
func NewPassiveFSM2(peer *Peer, con *net.TCPConn) *FSM2 {
	fsm := &FSM2{
		peer:          peer,
		eventCh:       make(chan int),
		con:           con,
		conCh:         make(chan net.Conn),
		msgRecvCh:     make(chan msgRecvMsg),
		msgRecvFailCh: make(chan msgRecvErr),
		stopMsgRecvCh: make(chan struct{}),
	}

	return fsm
}

// NewActiveFSM2 initiates a new passive FSM
func NewActiveFSM2(peer *Peer) *FSM2 {
	return &FSM2{
		peer:    peer,
		eventCh: make(chan int),
		active:  true,
		conCh:   make(chan net.Conn),
	}
}

func (fsm *FSM2) Cease() {

}

func (fsm *FSM2) startConnectRetryTimer() {
	fsm.connectRetryTimer = time.NewTimer(time.Second * fsm.connectRetryTime)
}

func (fsm *FSM2) resetConnectRetryTimer() {
	if !fsm.connectRetryTimer.Reset(time.Second * fsm.connectRetryTime) {
		<-fsm.connectRetryTimer.C
	}
}

func (fsm *FSM2) resetConnectRetryCounter() {
	fsm.connectRetryCounter = 0
}

func (fsm *FSM2) tcpConnect() {

}

func (fsm *FSM2) sendOpen() error {
	msg := packet.SerializeOpenMsg(&packet.BGPOpen{
		Version:       BGPVersion,
		AS:            uint16(fsm.peer.asn),
		HoldTime:      uint16(fsm.holdTimeConfigured),
		BGPIdentifier: fsm.server.routerID,
		OptParams:     fsm.peer.optOpenParams,
	})

	_, err := fsm.con.Write(msg)
	if err != nil {
		return fmt.Errorf("Unable to send OPEN message: %v", err)
	}

	return nil
}

func (fsm *FSM2) sendNotification(errorCode uint8, errorSubCode uint8) error {
	msg := packet.SerializeNotificationMsg(&packet.BGPNotification{})

	_, err := fsm.con.Write(msg)
	if err != nil {
		return fmt.Errorf("Unable to send NOTIFICATION message: %v", err)
	}

	return nil
}

func (fsm *FSM2) sendKeepalive() error {
	msg := packet.SerializeKeepaliveMsg()

	_, err := fsm.con.Write(msg)
	if err != nil {
		return fmt.Errorf("Unable to send KEEPALIVE message: %v", err)
	}

	return nil
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}
