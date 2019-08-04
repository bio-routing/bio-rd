package server

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/bio-routing/bio-rd/protocols/ospfv3/packet"
)

type NeighborState uint16

// neighbor state enum
const (
	NbrStateDown NeighborState = iota
	NbrStateAttempt
	NbrStateInit
	NbrState2Way
	NbrStateExStart
	NbrStateExchange
	NbrStateLoading
	NbrStateFull
)

type NeighborEvent uint16

// neighbor events
const (
	NbrEventHelloReceived NeighborEvent = iota
	NbrEventStart
	NbrEvent2WayReceived
	NbrEventNegotiationDone
	NbrEventExchangeDone
	NbrEventBadLSReq
	NbrEventLoadingDone
	NbrEventAdjOK
	NbrEventSeqNumMismatch
	NbrEvent1Way
	NbrEventKillNbr
	NbrEventInactivityTimer
	NbrEventLLDown
)

type neighborIf interface {
	Router
	GetConfig() InterfaceConfig

	sendHello(t time.Time, nbr *neighbor) error
	sendDatabaseDescription(nbr *neighbor, pl packet.DatabaseDescription) error
}

type NeighborAttributes struct {
	priority uint8
	address  net.IP
	options  packet.RouterOptions
	dr       packet.ID
	bdr      packet.ID
}

type neighbor struct {
	ifMgmt neighborIf
	events chan NeighborEvent

	// only changed by state machine
	state    NeighborState
	primary  bool
	ddSeqNum uint32
	lastDD   packet.DatabaseDescription

	id packet.ID
	// changed through SetAttributes()
	attrs  NeighborAttributes
	attrMu sync.Mutex

	rtmxLSAs    []packet.LSA
	dbSumLSAs   []packet.LSA
	requestLSAs []packet.LSA

	inactivityTimer *time.Timer
}

func match(a uint16, b uint16) uint64 {
	return (uint64(a) << 16) + uint64(b)
}

func matchN(state NeighborState, event NeighborEvent) uint64 {
	return match(uint16(state), uint16(event))
}

func newNeighbor(ifManager neighborIf, routerID packet.ID, prio uint8) *neighbor {
	neigh := &neighbor{
		ifMgmt: ifManager,
		events: make(chan NeighborEvent, 10),
		state:  NbrStateDown,
		id:     routerID,
		attrs: NeighborAttributes{
			priority: prio,
		},
	}
	return neigh
}

func (n *neighbor) RunEvent(event NeighborEvent) {
	n.events <- event
}

func (n *neighbor) IsDR() bool {
	return n.id == n.Attributes().dr
}

func (n *neighbor) IsBDR() bool {
	return n.id == n.Attributes().bdr
}

func (n *neighbor) Attributes() NeighborAttributes {
	n.attrMu.Lock()
	attrs := n.attrs
	n.attrMu.Unlock()
	return attrs
}

func (n *neighbor) SetAttributes(attrs NeighborAttributes) NeighborAttributes {
	n.attrMu.Lock()
	prevAttrs := n.attrs
	n.attrs = attrs
	n.attrMu.Unlock()
	return prevAttrs
}

func (n *neighbor) start(ctx context.Context) {
	go n.monitorEvents(ctx)
}

func (n *neighbor) monitorEvents(ctx context.Context) {
	for {
		var event NeighborEvent
		select {
		case <-ctx.Done():
			return
		case event = <-n.events:
		}

		n.state = n.runStateMachine(event)
	}
}

func (n *neighbor) resetTimer() {
	interval := n.ifMgmt.GetConfig().RouterDeadInterval

	if n.inactivityTimer == nil {
		n.inactivityTimer = time.AfterFunc(interval, func() {
			n.RunEvent(NbrEventInactivityTimer)
		})
		return
	}

	n.inactivityTimer.Reset(interval)
}

func (n *neighbor) runStateMachine(event NeighborEvent) NeighborState {
	// TODO: Neighbor State Machine: (https://tools.ietf.org/html/rfc2328#page-89)
	switch matchN(n.state, event) {
	case matchN(NbrStateDown, NbrEventStart):
		n.ifMgmt.sendHello(time.Now(), n)
		n.resetTimer()
		return NbrStateAttempt
	case matchN(NbrStateAttempt, NbrEventHelloReceived):
		n.resetTimer()
		return NbrStateInit
	case matchN(NbrStateDown, NbrEventHelloReceived):
		n.resetTimer()
		return NbrStateInit
	case matchN(NbrStateInit, NbrEvent2WayReceived):
		adjacent := shouldFormAdjacency(n.ifMgmt.GetConfig().IfType, n.ifMgmt, n)
		if !adjacent {
			return NbrState2Way
		}

		if n.ddSeqNum == 0 {
			n.ddSeqNum = uint32(time.Now().Unix())
		} else {
			n.ddSeqNum++
		}

		flags := packet.DBFlagInit & packet.DBFlagMore & packet.DBFlagMS
		n.ifMgmt.sendDatabaseDescription(n, packet.DatabaseDescription{
			DDSequenceNumber: n.ddSeqNum,
			DBFlags:          flags,
		})
		// todo: retransmit of database description

		return NbrStateExStart
	default:
		if n.state >= NbrStateInit && event == NbrEventHelloReceived {
			n.resetTimer()
			return n.state
		}
	}
	return n.state
}
