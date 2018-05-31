package server

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/bio-routing/bio-rd/routingtable"

	"github.com/bio-routing/bio-rd/config"
	tnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/packet"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBIn"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBOut"
	"github.com/bio-routing/bio-rd/routingtable/adjRIBOutAddPath"
	log "github.com/sirupsen/logrus"
	tomb "gopkg.in/tomb.v2"
)

const (
	// Administrative events
	ManualStart                               = 1
	ManualStop                                = 2
	AutomaticStart                            = 3
	ManualStartWithPassiveTcpEstablishment    = 4
	AutomaticStartWithPassiveTcpEstablishment = 5
	AutomaticStop                             = 8

	// Timer events
	ConnectRetryTimerExpires = 9
	HoldTimerExpires         = 10
	KeepaliveTimerExpires    = 11
)

const (
	Cease       = 0
	Idle        = 1
	Connect     = 2
	Active      = 3
	OpenSent    = 4
	OpenConfirm = 5
	Established = 6
)

type FSM struct {
	peer *Peer

	t           tomb.Tomb
	stateReason string
	state       int
	lastState   int
	eventCh     chan int

	con         *net.TCPConn
	con2        *net.TCPConn
	conCh       chan *net.TCPConn
	conErrCh    chan error
	initiateCon chan struct{}
	passive     bool

	local  net.IP
	remote net.IP

	localASN  uint16
	remoteASN uint16

	neighborID uint32
	routerID   uint32

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
	keepaliveTimer *time.Ticker

	msgRecvCh     chan msgRecvMsg
	msgRecvFailCh chan msgRecvErr
	stopMsgRecvCh chan struct{}

	adjRIBIn     *adjRIBIn.AdjRIBIn
	adjRIBOut    routingtable.RouteTableClient
	rib          routingtable.RouteTableClient
	updateSender routingtable.RouteTableClient

	capAddPathSend bool
	capAddPathRecv bool
}

type msgRecvMsg struct {
	msg []byte
	con *net.TCPConn
}

type msgRecvErr struct {
	err error
	con *net.TCPConn
}

func NewFSM(peer *Peer, c config.Peer, rib routingtable.RouteTableClient) *FSM {
	fsm := &FSM{
		peer:              peer,
		state:             Idle,
		passive:           true,
		connectRetryTime:  5,
		connectRetryTimer: time.NewTimer(time.Second * time.Duration(20)),

		msgRecvCh:     make(chan msgRecvMsg),
		msgRecvFailCh: make(chan msgRecvErr),
		stopMsgRecvCh: make(chan struct{}),

		holdTimeConfigured: time.Duration(c.HoldTimer),
		holdTimer:          time.NewTimer(0),

		keepaliveTime:  time.Duration(c.KeepAlive),
		keepaliveTimer: time.NewTicker(time.Duration(c.KeepAlive)),

		routerID: c.RouterID,
		remote:   c.PeerAddress,
		local:    c.LocalAddress,
		localASN: uint16(c.LocalAS),
		eventCh:  make(chan int),
		conCh:    make(chan *net.TCPConn),
		conErrCh: make(chan error), initiateCon: make(chan struct{}),

		rib: rib,
	}

	return fsm
}

func (fsm *FSM) disconnect() {
	if fsm.con != nil {
		fsm.con.Close()
		fsm.con = nil
	}
	if fsm.con2 != nil {
		fsm.con2.Close()
		fsm.con2 = nil
	}
}

func (fsm *FSM) changeState(new int, reason string) int {
	state := map[int]string{
		Cease:       "Cease",
		Idle:        "Idle",
		Connect:     "Connect",
		Active:      "Active",
		OpenSent:    "OpenSent",
		OpenConfirm: "OpenConfirm",
		Established: "Established",
	}

	log.WithFields(log.Fields{
		"peer":       fsm.remote.String(),
		"last_state": state[fsm.state],
		"new_state":  state[new],
		"reason":     reason,
	}).Info("FSM: Neighbor state change")

	fsm.lastState = fsm.state
	fsm.state = new
	fsm.stateReason = reason

	return fsm.state
}

func (fsm *FSM) activate() {
	fsm.eventCh <- ManualStart
}

func (fsm *FSM) Stop() error {
	fsm.eventCh <- ManualStop
	fsm.t.Kill(nil)
	return fsm.t.Wait()
}

func (fsm *FSM) start() {
	fsm.t.Go(fsm.main)
	fsm.t.Go(fsm.tcpConnector)
	return
}

func (fsm *FSM) main() error {
	next := fsm.idle()
	for {
		switch next {
		case Cease:
			fsm.t.Kill(fmt.Errorf("FSM is being stopped"))
			return nil
		case Idle:
			next = fsm.idle()
		case Connect:
			next = fsm.connect()
		case Active:
			next = fsm.active()
		case OpenSent:
			next = fsm.openSent()
		case OpenConfirm:
			next = fsm.openConfirm()
		case Established:
			next = fsm.established()
		}
	}
}

func (fsm *FSM) idle() int {
	if fsm.adjRIBOut != nil {
		fsm.rib.Unregister(fsm.adjRIBOut)
		fsm.adjRIBOut.Unregister(fsm.updateSender)
	}
	fsm.adjRIBIn = nil
	fsm.adjRIBOut = nil
	for {
		select {
		case c := <-fsm.conCh:
			c.Close()
			continue
		case e := <-fsm.eventCh:
			reason := ""
			switch e {
			case ManualStart:
				reason = "Received ManualStart event %d for %s peer"
			case AutomaticStart:
				reason = "Received AutomaticStart event %d for %s peer"
			default:
				continue
			}

			fsm.connectRetryCounter = 0
			fsm.startConnectRetryTimer()
			if fsm.passive {
				return fsm.changeState(Active, fmt.Sprintf(reason, "passive"))
			}
			fsm.tcpConnect()
			return fsm.changeState(Connect, fmt.Sprintf(reason, "active"))
		}

	}
}

func (fsm *FSM) tcpConnector() error {
	for {
		select {
		case <-fsm.initiateCon:
			c, err := net.DialTCP("tcp", &net.TCPAddr{IP: fsm.local}, &net.TCPAddr{IP: fsm.remote, Port: BGPPORT})
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
		case <-fsm.t.Dying():
			return nil
		}
	}
}

func (fsm *FSM) tcpConnect() {
	fsm.initiateCon <- struct{}{}
}

// connect state waits for a TCP connection to be fully established. Either the active or passive one.
func (fsm *FSM) connect() int {
	for {
		select {
		case e := <-fsm.eventCh:
			if e == ManualStop {
				fsm.connectRetryCounter = 0
				stopTimer(fsm.connectRetryTimer)
				return fsm.changeState(Idle, "Manual stop event")
			}
			continue
		case <-fsm.connectRetryTimer.C:
			fsm.resetConnectRetryTimer()
			fsm.tcpConnect()
			continue
		case c := <-fsm.conCh:
			fsm.con = c
			stopTimer(fsm.connectRetryTimer)
			return fsm.connectSendOpen()
		}
	}
}

func (fsm *FSM) connectSendOpen() int {
	err := fsm.sendOpen(fsm.con)
	if err != nil {
		stopTimer(fsm.connectRetryTimer)
		return fsm.changeState(Idle, fmt.Sprintf("Sending OPEN message failed: %v", err))
	}
	fsm.holdTimer = time.NewTimer(time.Minute * 4)
	return fsm.changeState(OpenSent, "Sent OPEN message")
}

// in the active state we wait for a passive TCP connection to be established
func (fsm *FSM) active() int {
	for {
		select {
		case e := <-fsm.eventCh:
			if e == ManualStop {
				fsm.disconnect()
				fsm.connectRetryCounter = 0
				stopTimer(fsm.connectRetryTimer)
				return fsm.changeState(Active, "Manual stop event")
			}
			continue
		case <-fsm.connectRetryTimer.C:
			fsm.resetConnectRetryTimer()
			fsm.tcpConnect()
			return fsm.changeState(Connect, "Connect retry timer expired")
		case c := <-fsm.conCh:
			fsm.con = c
			stopTimer(fsm.connectRetryTimer)
			return fsm.activeSendOpen()
		}
	}
}

func (fsm *FSM) activeSendOpen() int {
	err := fsm.sendOpen(fsm.con)
	if err != nil {
		fsm.resetConnectRetryTimer()
		fsm.connectRetryCounter++
		return fsm.changeState(Idle, fmt.Sprintf("Sending OPEN message failed: %v", err))
	}
	fsm.holdTimer = time.NewTimer(time.Minute * 4)
	return fsm.changeState(OpenSent, "Sent OPEN message")
}

func (fsm *FSM) msgReceiver(c *net.TCPConn) error {
	for {
		msg, err := recvMsg(c)
		if err != nil {
			fsm.msgRecvFailCh <- msgRecvErr{err: err, con: c}
			return nil

			/*select {
			case fsm.msgRecvFailCh <- msgRecvErr{err: err, con: c}:
				continue
			case <-time.NewTimer(time.Second * 60).C:
				return nil
			}*/
		}
		fsm.msgRecvCh <- msgRecvMsg{msg: msg, con: c}

		select {
		case <-fsm.stopMsgRecvCh:
			return nil
		default:
			continue
		}
	}
}

func (fsm *FSM) openSent() int {
	go fsm.msgReceiver(fsm.con)

	for {
		select {
		case e := <-fsm.eventCh:
			if e == ManualStop {
				sendNotification(fsm.con, packet.Cease, 0)
				stopTimer(fsm.connectRetryTimer)
				fsm.disconnect()
				fsm.connectRetryCounter = 0
				return fsm.changeState(Idle, "Manual stop event")
			}
			continue
		case <-fsm.holdTimer.C:
			sendNotification(fsm.con, packet.HoldTimeExpired, 0)
			stopTimer(fsm.connectRetryTimer)
			fsm.disconnect()
			fsm.connectRetryCounter++
			return fsm.changeState(Idle, "Holdtimer expired")
		case c := <-fsm.conCh: // 2nd connection coming in. Collision!
			if fsm.con2 != nil {
				log.WithFields(log.Fields{
					"peer":  fsm.remote.String(),
					"local": fsm.local.String(),
				}).Warningf("Received third connection from peer. Dropping new connection")
				c.Close()
				continue
			}

			err := fsm.sendOpen(c) // FIXME: Not sure if this is standard compliant
			if err != nil {
				c.Close()
				continue
			}
			fsm.con2 = c
			go fsm.msgReceiver(c)
			continue
		case recvMsg := <-fsm.msgRecvCh:
			msg, err := packet.Decode(bytes.NewBuffer(recvMsg.msg))
			if err != nil {
				switch bgperr := err.(type) {
				case packet.BGPError:
					sendNotification(fsm.con, bgperr.ErrorCode, bgperr.ErrorSubCode)
					sendNotification(fsm.con2, bgperr.ErrorCode, bgperr.ErrorSubCode)
				}
				stopTimer(fsm.connectRetryTimer)
				fsm.disconnect()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, fmt.Sprintf("Failed to decode BGP message: %v", err))
			}
			switch msg.Header.Type {
			case packet.NotificationMsg:
				nMsg := msg.Body.(*packet.BGPNotification)
				if nMsg.ErrorCode == packet.UnsupportedVersionNumber {
					stopTimer(fsm.connectRetryTimer)
					fsm.disconnect()
					return fsm.changeState(Idle, "Received NOTIFICATION")
				}

				if nMsg.ErrorCode == packet.Cease {
					// Was this connection to be closed anyways?
					if fsm.dumpCon(recvMsg.con) {
						continue
					}
				}
				stopTimer(fsm.connectRetryTimer)
				fsm.disconnect()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, "Received NOTIFICATION")
			case packet.OpenMsg:
				openMsg := msg.Body.(*packet.BGPOpen)
				fsm.neighborID = openMsg.BGPIdentifier
				fsm.resolveCollision()
				stopTimer(fsm.connectRetryTimer)
				err := fsm.sendKeepalive()
				if err != nil {
					return fsm.openSentTCPFail(err)
				}
				fsm.holdTime = time.Duration(math.Min(float64(fsm.holdTimeConfigured), float64(openMsg.HoldTime)))
				if fsm.holdTime != 0 {
					fsm.holdTimer.Reset(time.Second * fsm.holdTime)
					fsm.keepaliveTime = fsm.holdTime / 3
					if fsm.keepaliveTimer != nil {
						fsm.keepaliveTimer.Stop()
					}
					fsm.keepaliveTimer = time.NewTicker(fsm.keepaliveTime * time.Second)
				}

				fsm.processOpenOptions(openMsg.OptParams)

				return fsm.changeState(OpenConfirm, "Received OPEN message")
			default:
				sendNotification(fsm.con, packet.FiniteStateMachineError, 0)
				stopTimer(fsm.connectRetryTimer)
				fsm.con.Close()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, "FSM Error")
			}
		case err := <-fsm.msgRecvFailCh:
			if err.con == fsm.con && fsm.con2 != nil {
				fsm.con.Close()
				fsm.con = fsm.con2
				fsm.con2 = nil
				continue
			}

			if err.con == fsm.con2 {
				fsm.con2.Close()
				fsm.con2 = nil
				continue
			}
			return fsm.openSentTCPFail(err.err)
		}
	}
}

func (fsm *FSM) processOpenOptions(optParams []packet.OptParam) {
	for _, optParam := range optParams {
		if optParam.Type != packet.CapabilitiesParamType {
			continue
		}

		fsm.processCapabilities(optParam.Value.(packet.Capabilities))
	}
}

func (fsm *FSM) processCapabilities(caps packet.Capabilities) {
	for _, cap := range caps {
		fsm.processCapability(cap)
	}
}

func (fsm *FSM) processCapability(cap packet.Capability) {
	switch cap.Code {
	case packet.AddPathCapabilityCode:
		fsm.processAddPathCapability(cap.Value.(packet.AddPathCapability))

	}
}

func (fsm *FSM) processAddPathCapability(addPathCap packet.AddPathCapability) {
	if addPathCap.AFI != 1 {
		return
	}
	if addPathCap.SAFI != 1 {
		return
	}
	switch addPathCap.SendReceive {
	case packet.AddPathReceive:
		if !fsm.peer.addPathSend.BestOnly {
			fsm.capAddPathSend = true
		}
	case packet.AddPathSend:
		if fsm.peer.addPathRecv {
			fsm.capAddPathRecv = true
		}
	case packet.AddPathSendReceive:
		if !fsm.peer.addPathSend.BestOnly {
			fsm.capAddPathSend = true
		}
		if fsm.peer.addPathRecv {
			fsm.capAddPathRecv = true
		}
	}
}

func (fsm *FSM) openSentTCPFail(err error) int {
	fsm.con.Close()
	fsm.resetConnectRetryTimer()
	return fsm.changeState(Active, fmt.Sprintf("TCP failure: %v", err))
}

func (fsm *FSM) dumpCon(c *net.TCPConn) bool {
	p := fsm.isPassive(c)
	if fsm.routerID > fsm.neighborID {
		return p
	}
	return !p
}

func (fsm *FSM) resolveCollision() {
	if fsm.con2 == nil {
		return
	}

	if fsm.routerID > fsm.neighborID {
		// Terminate passive connection
		if fsm.isPassive(fsm.con) {
			dumpCon(fsm.con)
			fsm.con = fsm.con2
			return
		}
		if fsm.isPassive(fsm.con2) {
			dumpCon(fsm.con2)
			return
		}
		return
	}

	// Terminate active connection
	if !fsm.isPassive(fsm.con) {
		dumpCon(fsm.con)
		fsm.con = fsm.con2
		return
	}
	if !fsm.isPassive(fsm.con2) {
		dumpCon(fsm.con2)
		fsm.con2.Close()
		fsm.con2 = nil
		return
	}
}

func dumpCon(c *net.TCPConn) {
	sendNotification(c, packet.Cease, packet.ConnectionCollisionResolution)
	c.Close()
}

func (fsm *FSM) isPassive(c *net.TCPConn) bool {
	if c.LocalAddr().String() == fmt.Sprintf("%s:179", fsm.local.String()) {
		return true
	}
	return false
}

func (fsm *FSM) openConfirm() int {
	for {
		select {
		case e := <-fsm.eventCh:
			if e == ManualStop { // Event 2
				sendNotification(fsm.con, packet.Cease, 0)
				stopTimer(fsm.connectRetryTimer)
				fsm.disconnect()
				fsm.connectRetryCounter = 0
				return fsm.changeState(Idle, "Manual stop event")
			}
			continue
		case <-fsm.holdTimer.C:
			sendNotification(fsm.con, packet.HoldTimeExpired, 0)
			stopTimer(fsm.connectRetryTimer)
			fsm.disconnect()
			fsm.connectRetryCounter++
			return fsm.changeState(Idle, "Holdtimer expired")
		case <-fsm.keepaliveTimer.C:
			err := fsm.sendKeepalive()
			if err != nil {
				stopTimer(fsm.connectRetryTimer)
				fsm.disconnect()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, fmt.Sprintf("Failed to send keepalive: %v", err))
			}
			continue
		case c := <-fsm.conCh:
			if fsm.con2 != nil {
				log.WithFields(log.Fields{
					"peer":  fsm.remote.String(),
					"local": fsm.local.String(),
				}).Warningf("Received third connection from peer. Dropping new connection")
				c.Close()
				continue
			}

			err := fsm.sendOpen(c) // FIXME: Not sure if this is standard compliant
			if err != nil {
				c.Close()
				continue
			}
			fsm.con2 = c
			go fsm.msgReceiver(c)
			continue
		case recvMsg := <-fsm.msgRecvCh:
			msg, err := packet.Decode(bytes.NewBuffer(recvMsg.msg))
			if err != nil {
				log.WithError(err).Errorf("Failed to decode BGP message %v\n", recvMsg.msg)
				switch bgperr := err.(type) {
				case packet.BGPError:
					sendNotification(fsm.con, bgperr.ErrorCode, bgperr.ErrorSubCode)
					sendNotification(fsm.con2, bgperr.ErrorCode, bgperr.ErrorSubCode)
				}
				stopTimer(fsm.connectRetryTimer)
				fsm.disconnect()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, "Failed to decode BGP message")
			}

			switch msg.Header.Type {
			case packet.NotificationMsg:
				nMsg := msg.Body.(*packet.BGPNotification)
				if nMsg.ErrorCode == packet.UnsupportedVersionNumber {
					stopTimer(fsm.connectRetryTimer)
					fsm.con.Close()
					return fsm.changeState(Idle, "Received NOTIFICATION")
				}

				if nMsg.ErrorCode == packet.Cease {
					// Was this connection to be closed anyways?
					if fsm.dumpCon(recvMsg.con) {
						continue
					}
				}

				return fsm.openConfirmTCPFail(fmt.Errorf("NOTIFICATION received"))
			case packet.KeepaliveMsg:
				fsm.holdTimer.Reset(time.Second * fsm.holdTime)
				return fsm.changeState(Established, "Received KEEPALIVE")
			case packet.OpenMsg:
				openMsg := msg.Body.(*packet.BGPOpen)
				fsm.neighborID = openMsg.BGPIdentifier
				fsm.resolveCollision()
			default:
				sendNotification(fsm.con, packet.FiniteStateMachineError, 0)
				stopTimer(fsm.connectRetryTimer)
				fsm.con.Close()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, "FSM Error")
			}
		case err := <-fsm.msgRecvFailCh:
			if err.con == fsm.con && fsm.con2 != nil {
				fsm.con.Close()
				fsm.con = fsm.con2
				fsm.con2 = nil
				continue
			}

			if err.con == fsm.con2 {
				fsm.con2.Close()
				fsm.con2 = nil
				continue
			}
			return fsm.openConfirmTCPFail(err.err)
		}
	}
}

func (fsm *FSM) openConfirmTCPFail(err error) int {
	fsm.con.Close()
	fsm.resetConnectRetryTimer()
	fsm.connectRetryCounter++
	return fsm.changeState(Idle, fmt.Sprintf("Failure: %v", err))
}

func (fsm *FSM) established() int {
	fsm.adjRIBIn = adjRIBIn.New()
	fsm.adjRIBIn.Register(fsm.rib)

	n := &routingtable.Neighbor{
		Type:    route.BGPPathType,
		Address: tnet.IPv4ToUint32(fsm.remote),
	}

	clientOptions := routingtable.ClientOptions{
		BestOnly: true,
	}
	if fsm.capAddPathSend {
		fsm.updateSender = newUpdateSenderAddPath(fsm)
		fsm.adjRIBOut = adjRIBOutAddPath.New(n)
		clientOptions = fsm.peer.addPathSend
	} else {
		fsm.updateSender = newUpdateSender(fsm)
		fsm.adjRIBOut = adjRIBOut.New(n)
	}

	fsm.adjRIBOut.Register(fsm.updateSender)
	fsm.rib.RegisterWithOptions(fsm.adjRIBOut, clientOptions)

	/*go func() {
		for {
			if fsm.adjRIBOut == nil {
				return
			}
			fmt.Printf("ADJ-RIB-OUT: %s\n", fsm.remote.String())
			fmt.Print(fsm.adjRIBOut.Print())
			time.Sleep(time.Second * 11)
		}
	}()*/

	for {
		log.Debug("Iterate established loop.")
		select {
		case e := <-fsm.eventCh:
			if e == ManualStop { // Event 2
				sendNotification(fsm.con, packet.Cease, 0)
				stopTimer(fsm.connectRetryTimer)
				fsm.con.Close()
				fsm.connectRetryCounter = 0
				return fsm.changeState(Idle, "Manual stop event")
			}
			if e == AutomaticStop { // Event 8
				sendNotification(fsm.con, packet.Cease, 0)
				stopTimer(fsm.connectRetryTimer)
				fsm.con.Close()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, "Automatic stop event")
			}
			continue
		case <-fsm.holdTimer.C:
			sendNotification(fsm.con, packet.HoldTimeExpired, 0)
			stopTimer(fsm.connectRetryTimer)
			fsm.con.Close()
			fsm.connectRetryCounter++
			return fsm.changeState(Idle, "Holdtimer expired")
		case <-fsm.keepaliveTimer.C:

			err := fsm.sendKeepalive()
			if err != nil {
				stopTimer(fsm.connectRetryTimer)
				fsm.con.Close()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, fmt.Sprintf("Failed to send keepalive: %v", err))
			}
			continue
		case c := <-fsm.conCh:
			c.Close()
			continue
		case recvMsg := <-fsm.msgRecvCh:
			msg, err := packet.Decode(bytes.NewBuffer(recvMsg.msg))
			if err != nil {
				log.WithError(err).Errorf("Failed to decode BGP message %v\n", recvMsg.msg)
				switch bgperr := err.(type) {
				case packet.BGPError:
					sendNotification(fsm.con, bgperr.ErrorCode, bgperr.ErrorSubCode)
				}
				stopTimer(fsm.connectRetryTimer)
				fsm.con.Close()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, "Failed to decode BGP message")
			}
			switch msg.Header.Type {
			case packet.NotificationMsg:
				stopTimer(fsm.connectRetryTimer)
				fsm.con.Close()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, "Received NOTIFICATION")
			case packet.UpdateMsg:
				if fsm.holdTime != 0 {
					fsm.holdTimer.Reset(time.Second * fsm.holdTime)
				}

				u := msg.Body.(*packet.BGPUpdate)

				for r := u.WithdrawnRoutes; r != nil; r = r.Next {
					pfx := tnet.NewPfx(r.IP, r.Pfxlen)
					log.WithField("Prefix", pfx.String()).Debug("LPM: Removing prefix")
					fsm.adjRIBIn.RemovePath(pfx, nil)
				}

				for r := u.NLRI; r != nil; r = r.Next {
					pfx := tnet.NewPfx(r.IP, r.Pfxlen)
					log.WithField("Prefix", pfx.String()).Debug("LPM: Adding prefix")
					path := &route.Path{
						Type: route.BGPPathType,
						BGPPath: &route.BGPPath{
							Source: tnet.IPv4ToUint32(fsm.remote),
						},
					}

					for pa := u.PathAttributes; pa != nil; pa = pa.Next {
						switch pa.TypeCode {
						case packet.OriginAttr:
							path.BGPPath.Origin = pa.Value.(uint8)
						case packet.LocalPrefAttr:
							path.BGPPath.LocalPref = pa.Value.(uint32)
						case packet.MEDAttr:
							path.BGPPath.MED = pa.Value.(uint32)
						case packet.NextHopAttr:
							log.WithField("NextHop", pa.Value.(uint32)).Debug("RECEIVED NEXT_HOP")
							path.BGPPath.NextHop = pa.Value.(uint32)
						case packet.ASPathAttr:
							path.BGPPath.ASPath = pa.ASPathString()
							path.BGPPath.ASPathLen = pa.ASPathLen()
						case packet.CommunitiesAttr:
							path.BGPPath.Communities = pa.CommunityString()
						case packet.LargeCommunityAttr:
							path.BGPPath.LargeCommunities = pa.LargeCommunityString()
						}
					}
					fsm.adjRIBIn.AddPath(pfx, path)
				}

				continue
			case packet.KeepaliveMsg:
				if fsm.holdTime != 0 {
					fsm.holdTimer.Reset(time.Second * fsm.holdTime)
				}
				continue
			case packet.OpenMsg:
				if fsm.con2 != nil {
					sendNotification(fsm.con2, packet.Cease, packet.ConnectionCollisionResolution)
					fsm.con2.Close()
					fsm.con2 = nil
					continue
				}
				sendNotification(fsm.con, packet.FiniteStateMachineError, 0)
				stopTimer(fsm.connectRetryTimer)
				fsm.con.Close()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, "FSM Error")
			default:
				sendNotification(fsm.con, packet.FiniteStateMachineError, 0)
				stopTimer(fsm.connectRetryTimer)
				fsm.con.Close()
				fsm.connectRetryCounter++
				return fsm.changeState(Idle, "FSM Error")
			}
		case err := <-fsm.msgRecvFailCh:
			if err.con == fsm.con && fsm.con2 != nil {
				fsm.con.Close()
				fsm.con = fsm.con2
				fsm.con2 = nil
				continue
			}

			if err.con == fsm.con2 {
				fsm.con2.Close()
				fsm.con2 = nil
				continue
			}
			return fsm.openConfirmTCPFail(err.err)
		}
	}
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func (fsm *FSM) startConnectRetryTimer() {
	fsm.connectRetryTimer = time.NewTimer(time.Second * fsm.connectRetryTime)
}

func (fsm *FSM) resetConnectRetryTimer() {
	if !fsm.connectRetryTimer.Reset(time.Second * fsm.connectRetryTime) {
		<-fsm.connectRetryTimer.C
	}
}

func (fsm *FSM) resetDelayOpenTimer() {
	if !fsm.delayOpenTimer.Reset(time.Second * fsm.delayOpenTime) {
		<-fsm.delayOpenTimer.C
	}
}

func (fsm *FSM) sendKeepalive() error {

	msg := packet.SerializeKeepaliveMsg()
	_, err := fsm.con.Write(msg)
	log.WithError(err).Debug("Send keepalive")
	if err != nil {
		return fmt.Errorf("Unable to send KEEPALIVE message: %v", err)
	}

	return nil
}

func (fsm *FSM) sendOpen(c *net.TCPConn) error {
	msg := packet.SerializeOpenMsg(&packet.BGPOpen{
		Version:       BGPVersion,
		AS:            fsm.localASN,
		HoldTime:      uint16(fsm.holdTimeConfigured),
		BGPIdentifier: fsm.routerID,
		OptParams:     fsm.peer.optOpenParams,
	})

	_, err := c.Write(msg)
	if err != nil {
		return fmt.Errorf("Unable to send OPEN message: %v", err)
	}

	return nil
}

func sendNotification(c *net.TCPConn, errorCode uint8, errorSubCode uint8) error {
	if c == nil {
		return fmt.Errorf("connection is nil")
	}

	msg := packet.SerializeNotificationMsg(&packet.BGPNotification{})

	_, err := c.Write(msg)
	if err != nil {
		return fmt.Errorf("Unable to send NOTIFICATION message: %v", err)
	}

	return nil
}
