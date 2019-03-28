package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/pkg/errors"
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
	stateNameIdle                             = "idle"
	stateNameConnect                          = "connect"
	stateNameActive                           = "active"
	stateNameOpenSent                         = "openSent"
	stateNameOpenConfirm                      = "openConfirm"
	stateNameEstablished                      = "established"
	stateNameCease                            = "cease"
)

type state interface {
	run() (state, string)
}

// FSM implements the BGP finite state machine (RFC4271)
type FSM struct {
	isBMP       bool
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

	holdTime              time.Duration
	lastUpdateOrKeepalive time.Time

	keepaliveTime  time.Duration
	keepaliveTimer *time.Timer

	msgRecvCh     chan []byte
	msgRecvFailCh chan error
	stopMsgRecvCh chan struct{}

	local net.IP

	ribsInitialized bool
	ipv4Unicast     *fsmAddressFamily
	ipv6Unicast     *fsmAddressFamily

	supports4OctetASN bool

	neighborID uint32
	state      state
	stateMu    sync.RWMutex
	reason     string
	active     bool

	establishedTime time.Time
	counters        fsmCounters

	connectionCancelFunc context.CancelFunc
}

// NewPassiveFSM initiates a new passive FSM
func NewPassiveFSM(peer *peer, con *net.TCPConn) *FSM {
	fsm := newFSM(peer)
	fsm.con = con
	fsm.state = newIdleState(fsm)
	return fsm
}

// NewActiveFSM initiates a new passive FSM
func NewActiveFSM(peer *peer) *FSM {
	fsm := newFSM(peer)
	fsm.active = true
	fsm.state = newIdleState(fsm)
	return fsm
}

func newFSM(peer *peer) *FSM {
	f := &FSM{
		connectRetryTime: time.Minute,
		peer:             peer,
		eventCh:          make(chan int),
		conCh:            make(chan net.Conn),
		conErrCh:         make(chan error),
		initiateCon:      make(chan struct{}),
		msgRecvCh:        make(chan []byte),
		msgRecvFailCh:    make(chan error),
		stopMsgRecvCh:    make(chan struct{}),
		counters:         fsmCounters{},
	}

	if peer.ipv4 != nil {
		f.ipv4Unicast = newFSMAddressFamily(packet.IPv4AFI, packet.UnicastSAFI, peer.ipv4, f)
	}

	if peer.ipv6 != nil {
		f.ipv6Unicast = newFSMAddressFamily(packet.IPv6AFI, packet.UnicastSAFI, peer.ipv6, f)
	}

	return f
}

func (fsm *FSM) updateLastUpdateOrKeepalive() {
	fsm.lastUpdateOrKeepalive = time.Now()
}

func (fsm *FSM) addressFamily(afi uint16, safi uint8) *fsmAddressFamily {
	if safi != packet.UnicastSAFI {
		return nil
	}

	switch afi {
	case packet.IPv4AFI:
		return fsm.ipv4Unicast
	case packet.IPv6AFI:
		return fsm.ipv6Unicast
	default:
		return nil
	}
}

func (fsm *FSM) start() {
	ctx, cancel := context.WithCancel(context.Background())
	fsm.connectionCancelFunc = cancel

	go fsm.run()
	go fsm.tcpConnector(ctx)
	return
}

func (fsm *FSM) activate() {
	fsm.eventCh <- AutomaticStart
}

func (fsm *FSM) run() {
	defer fsm.cancelRunningGoRoutines()

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

		if newState == stateNameCease {
			return
		}

		if oldState != newState && newState == stateNameEstablished {
			fsm.establishedTime = time.Now()
		}

		fsm.stateMu.Lock()
		fsm.state = next
		fsm.stateMu.Unlock()

		next, reason = fsm.state.run()
	}
}

func (fsm *FSM) cancelRunningGoRoutines() {
	if fsm.connectionCancelFunc != nil {
		fsm.connectionCancelFunc()
	}
}

func stateName(s state) string {
	switch s.(type) {
	case *idleState:
		return stateNameIdle
	case *connectState:
		return stateNameConnect
	case *activeState:
		return stateNameActive
	case *openSentState:
		return stateNameOpenSent
	case *openConfirmState:
		return stateNameOpenConfirm
	case *establishedState:
		return stateNameEstablished
	case *ceaseState:
		return stateNameCease
	default:
		panic(fmt.Sprintf("Unknown state: %v", s))
	}
}

func (fsm *FSM) cease() {
	fsm.eventCh <- Cease
}

func (fsm *FSM) tcpConnector(ctx context.Context) {
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
			case <-ctx.Done():
				return
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

func (fsm *FSM) decodeOptions() *packet.DecodeOptions {
	return &packet.DecodeOptions{
		Use32BitASN: fsm.supports4OctetASN,
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
		return errors.Wrap(err, "Unable to send OPEN message")
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
		return errors.Wrap(err, "Unable to send NOTIFICATION message")
	}

	return nil
}

func (fsm *FSM) sendKeepalive() error {
	msg := packet.SerializeKeepaliveMsg()

	_, err := fsm.con.Write(msg)
	if err != nil {
		return errors.Wrap(err, "Unable to send KEEPALIVE message")
	}

	return nil
}

func recvMsg(c net.Conn) (msg []byte, err error) {
	buffer := make([]byte, packet.MaxLen)
	_, err = io.ReadFull(c, buffer[0:packet.MinLen])
	if err != nil {
		return nil, errors.Wrap(err, "Read failed")
	}

	l := int(buffer[16])*256 + int(buffer[17])
	toRead := l
	_, err = io.ReadFull(c, buffer[packet.MinLen:toRead])
	if err != nil {
		return nil, errors.Wrap(err, "Read failed")
	}

	return buffer, nil
}

func stopTimer(t *time.Timer) {
	if t == nil {
		return
	}

	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}
