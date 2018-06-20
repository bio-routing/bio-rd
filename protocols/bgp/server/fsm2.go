package server

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
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

// FSM2 implements the BGP finite state machine (RFC4271)
type FSM2 struct {
	peer        *Peer
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

	capAddPathSend bool
	capAddPathRecv bool

	local net.IP
	//remote net.IP

	ribsInitialized bool
	adjRIBIn        routingtable.RouteTableClient
	adjRIBOut       routingtable.RouteTableClient
	rib             *locRIB.LocRIB
	updateSender    routingtable.RouteTableClient

	neighborID uint32
	state      state
	stateMu    sync.RWMutex
	reason     string
	active     bool
}

// NewPassiveFSM2 initiates a new passive FSM
func NewPassiveFSM2(peer *Peer, con *net.TCPConn) *FSM2 {
	fmt.Printf("NewPassiveFSM2\n")
	fsm := newFSM2(peer)
	fsm.con = con
	fsm.state = newIdleState(fsm)
	return fsm
}

// NewActiveFSM2 initiates a new passive FSM
func NewActiveFSM2(peer *Peer) *FSM2 {
	fmt.Printf("NewActiveFSM2\n")
	fsm := newFSM2(peer)
	fsm.active = true
	fsm.state = newIdleState(fsm)
	return fsm
}

func newFSM2(peer *Peer) *FSM2 {
	return &FSM2{
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
	}
}

func (fsm *FSM2) start() {
	go fsm.run()
	go fsm.tcpConnector()
	return
}

func (fsm *FSM2) activate() {
	fsm.eventCh <- AutomaticStart
}

func (fsm *FSM2) run() {
	//fmt.Printf("Starting FSM\n")
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

		//fmt.Printf("Aquiring lock...\n")
		fsm.stateMu.Lock()
		fsm.state = next
		//fmt.Printf("Releasing lock...\n")
		fsm.stateMu.Unlock()

		//fmt.Printf("Running new state\n")
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

func (fsm *FSM2) cease() {
	fsm.eventCh <- Cease
}

func (fsm *FSM2) tcpConnector() error {
	fmt.Printf("TCP CONNECTOR STARTED\n")
	for {
		//fmt.Printf("READING FROM fsm.initiateCon\n")
		select {
		case <-fsm.initiateCon:
			fmt.Printf("Initiating connection to %s\n", fsm.peer.addr.String())
			c, err := net.DialTCP("tcp", &net.TCPAddr{IP: fsm.local}, &net.TCPAddr{IP: fsm.peer.addr, Port: BGPPORT})
			if err != nil {
				select {
				case fsm.conErrCh <- err:
					continue
				case <-time.NewTimer(time.Second * 30).C:
					continue
				}
			}

			//fmt.Printf("GOT CONNECTION!\n")
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

func (fsm *FSM2) tcpConnect() {
	fsm.initiateCon <- struct{}{}
}

func (fsm *FSM2) msgReceiver() error {
	for {
		msg, err := recvMsg(fsm.con)
		if err != nil {
			fsm.msgRecvFailCh <- err
			return nil

			/*select {
			case fsm.msgRecvFailCh <- msgRecvErr{err: err, con: c}:
				continue
			case <-time.NewTimer(time.Second * 60).C:
				return nil
			}*/
		}
		fmt.Printf("Message received for %s: %v\n", fsm.con.RemoteAddr().String(), msg[18])
		fsm.msgRecvCh <- msg
	}
}

func (fsm *FSM2) startConnectRetryTimer() {
	fmt.Printf("Initializing connectRetryTimer: %d\n", fsm.connectRetryTime)
	fsm.connectRetryTimer = time.NewTimer(fsm.connectRetryTime)
}

func (fsm *FSM2) resetConnectRetryTimer() {
	if !fsm.connectRetryTimer.Reset(fsm.connectRetryTime) {
		<-fsm.connectRetryTimer.C
	}
}

func (fsm *FSM2) resetConnectRetryCounter() {
	fsm.connectRetryCounter = 0
}

func (fsm *FSM2) sendOpen() error {
	fmt.Printf("Sending OPEN Message to %s\n", fsm.con.RemoteAddr().String())

	msg := packet.SerializeOpenMsg(&packet.BGPOpen{
		Version:       BGPVersion,
		AS:            uint16(fsm.peer.localASN),
		HoldTime:      uint16(fsm.peer.holdTime / time.Second),
		BGPIdentifier: fsm.peer.server.routerID,
		OptParams:     fsm.peer.optOpenParams,
	})

	_, err := fsm.con.Write(msg)
	if err != nil {
		return fmt.Errorf("Unable to send OPEN message: %v", err)
	}

	return nil
}

func (fsm *FSM2) sendNotification(errorCode uint8, errorSubCode uint8) error {
	fmt.Printf("Sending NOTIFICATION Message to %s\n", fsm.con.RemoteAddr().String())
	msg := packet.SerializeNotificationMsg(&packet.BGPNotification{})

	_, err := fsm.con.Write(msg)
	if err != nil {
		return fmt.Errorf("Unable to send NOTIFICATION message: %v", err)
	}

	return nil
}

func (fsm *FSM2) sendKeepalive() error {
	fmt.Printf("Sending KEEPALIVE to %s\n", fsm.con.RemoteAddr().String())
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
