package server

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	log "github.com/sirupsen/logrus"
)

const (
	// Administrative events
	ManualStart                               = 1
	ManualStop                                = 2
	AutomaticStart                            = 3
	ManualStartWithPassiveTcpEstablishment    = 4
	AutomaticStartWithPassiveTcpEstablishment = 5
	AutomaticStop                             = 8
	Cease                                     = 100
)

type state interface {
	run() (state, string)
}

// FSM implements the BGP finite state machine (RFC4271)
type FSM struct {
	peer        *peer
	eventCh     chan int
	con         net.Conn
	conCh       chan net.Conn
	initiateCon chan struct{}
	conErrCh    chan error

	delayOpen      bool
	delayOpenTime  time.Duration
	delayOpenTimer *time.Timer

	connectRetryTime    time.Duration
	connectRetryTimer   *time.Timer
	connectRetryCounter int

	holdTime  time.Duration
	holdTimer *time.Timer

	keepaliveTime  time.Duration
	keepaliveTimer *time.Timer

	msgRecvCh     chan []byte
	msgRecvFailCh chan error
	stopMsgRecvCh chan struct{}

	options *types.Options

	local net.IP

	ribsInitialized bool
	adjRIBIn        routingtable.RouteTableClient
	adjRIBOut       routingtable.RouteTableClient
	rib             *locRIB.LocRIB
	updateSender    *UpdateSender

	neighborID uint32
	state      state
	stateMu    sync.RWMutex
	reason     string
	active     bool
}

// NewPassiveFSM2 initiates a new passive FSM
func NewPassiveFSM2(peer *peer, con *net.TCPConn) *FSM {
	fsm := newFSM2(peer)
	fsm.con = con
	fsm.state = newIdleState(fsm)
	return fsm
}

// NewActiveFSM2 initiates a new passive FSM
func NewActiveFSM2(peer *peer) *FSM {
	fsm := newFSM2(peer)
	fsm.active = true
	fsm.state = newIdleState(fsm)
	return fsm
}

func newFSM2(peer *peer) *FSM {
	return &FSM{
		connectRetryTime: time.Minute,
		peer:             peer,
		eventCh:          make(chan int),
		conCh:            make(chan net.Conn),
		conErrCh:         make(chan error),
		initiateCon:      make(chan struct{}),
		msgRecvCh:        make(chan []byte),
		msgRecvFailCh:    make(chan error),
		stopMsgRecvCh:    make(chan struct{}),
		rib:              peer.rib,
		options:          &types.Options{},
	}
}

func (fsm *FSM) start() {
	go fsm.run()
	go fsm.tcpConnector()
	return
}

func (fsm *FSM) activate() {
	fsm.eventCh <- AutomaticStart
}

func (fsm *FSM) run() {
	next, reason := fsm.state.run()
	for {
		newState := stateName(next)
		oldState := stateName(fsm.state)

		if oldState != newState {
			log.WithFields(log.Fields{
				"peer":       fsm.peer.addr.String(),
				"last_state": oldState,
				"new_state":  newState,
				"reason":     reason,
			}).Info("FSM: Neighbor state change")
		}

		if newState == "cease" {
			return
		}

		fsm.stateMu.Lock()
		fsm.state = next
		fsm.stateMu.Unlock()

		next, reason = fsm.state.run()
	}
}

func stateName(s state) string {
	switch s.(type) {
	case *idleState:
		return "idle"
	case *connectState:
		return "connect"
	case *activeState:
		return "active"
	case *openSentState:
		return "openSent"
	case *openConfirmState:
		return "openConfirm"
	case *establishedState:
		return "established"
	case *ceaseState:
		return "cease"
	default:
		panic(fmt.Sprintf("Unknown state: %v", s))
	}
}

func (fsm *FSM) cease() {
	fsm.eventCh <- Cease
}

func (fsm *FSM) tcpConnector() error {
	for {
		select {
		case <-fsm.initiateCon:
			c, err := net.DialTCP("tcp", &net.TCPAddr{IP: fsm.local}, &net.TCPAddr{IP: fsm.peer.addr.ToNetIP(), Port: BGPPORT})
			if err != nil {
				select {
				case fsm.conErrCh <- err:
					continue
				case <-time.NewTimer(time.Second * 30).C:
					continue
				}
			}

			select {
			case fsm.conCh <- c:
				continue
			case <-time.NewTimer(time.Second * 30).C:
				c.Close()
				continue
			}
		}
	}
}

func (fsm *FSM) tcpConnect() {
	fsm.initiateCon <- struct{}{}
}

func (fsm *FSM) msgReceiver() error {
	for {
		msg, err := recvMsg(fsm.con)
		if err != nil {
			fsm.msgRecvFailCh <- err
			return nil
		}
		fsm.msgRecvCh <- msg
	}
}

func (fsm *FSM) startConnectRetryTimer() {
	fsm.connectRetryTimer = time.NewTimer(fsm.connectRetryTime)
}

func (fsm *FSM) resetConnectRetryTimer() {
	if !fsm.connectRetryTimer.Reset(fsm.connectRetryTime) {
		<-fsm.connectRetryTimer.C
	}
}

func (fsm *FSM) resetConnectRetryCounter() {
	fsm.connectRetryCounter = 0
}

func (fsm *FSM) sendOpen() error {
	msg := packet.SerializeOpenMsg(fsm.openMessage())

	_, err := fsm.con.Write(msg)
	if err != nil {
		return fmt.Errorf("Unable to send OPEN message: %v", err)
	}

	return nil
}

func (fsm *FSM) openMessage() *packet.BGPOpen {
	return &packet.BGPOpen{
		Version:       BGPVersion,
		ASN:           fsm.local16BitASN(),
		HoldTime:      uint16(fsm.peer.holdTime / time.Second),
		BGPIdentifier: fsm.peer.routerID,
		OptParams:     fsm.peer.optOpenParams,
	}
}

func (fsm *FSM) local16BitASN() uint16 {
	if fsm.peer.localASN > uint32(^uint16(0)) {
		return packet.ASTransASN
	}

	return uint16(fsm.peer.localASN)
}

func (fsm *FSM) sendNotification(errorCode uint8, errorSubCode uint8) error {
	msg := packet.SerializeNotificationMsg(&packet.BGPNotification{})

	_, err := fsm.con.Write(msg)
	if err != nil {
		return fmt.Errorf("Unable to send NOTIFICATION message: %v", err)
	}

	return nil
}

func (fsm *FSM) sendKeepalive() error {
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
