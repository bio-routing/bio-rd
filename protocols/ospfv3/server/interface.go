package server

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/bio-routing/bio-rd/protocols/ospfv3/packet"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/ipv6"
)

type InterfaceType uint64

// interface type enum
const (
	IfTypePointToPoint InterfaceType = iota + 1
	IfTypeBroadcast
	IfTypeNBMA
	IfTypePointToMultipoint
	IfTypeVirtualLink
)

func guessIfType(intf *net.Interface) (InterfaceType, error) {
	if (intf.Flags & net.FlagPointToPoint) != 0 {
		return IfTypePointToPoint, nil
	}
	if (intf.Flags & net.FlagBroadcast & net.FlagMulticast) != 0 {
		return IfTypeBroadcast, nil
	}
	return 0, errors.New("No interface type given on link without broadcast or multicast capability")
}

type InterfaceState uint64

// interface state enum
const (
	IfStateDown InterfaceState = iota
	IfStateLoopback
	IfStateWaiting
	IfStatePTP
	IfStateDROther
	IfStateBackup
	IfStateDR
)

type InterfaceEvent uint64

// interface events
const (
	IfEventInterfaceUp InterfaceEvent = iota
	IfEventWaitTimer
	IfEventBackupSeen
	IfEventNeighborChange
	IfEventLoopInd
	IfEventUnloopInd
	IfEventInterfaceDown
)

type InterfaceConfig struct {
	IfType InterfaceType

	HelloInterval      time.Duration
	RouterDeadInterval time.Duration
	InfTransDelay      time.Duration
	RouterPriority     uint8
	IfCost             uint16
	RxmtInterval       time.Duration
}

func (conf InterfaceConfig) ApplyDefaults() InterfaceConfig {
	if conf.HelloInterval == 0 {
		conf.HelloInterval = 10 * time.Second
	}
	if conf.RouterDeadInterval == 0 {
		conf.RouterDeadInterval = 4 * conf.HelloInterval
	}
	if conf.InfTransDelay == 0 {
		conf.InfTransDelay = 1 * time.Second
	}
	if conf.RxmtInterval == 0 {
		conf.RxmtInterval = 5 * time.Second
	}
	return conf
}

func (conf InterfaceConfig) getInstanceID() uint8 {
	// todo: support multiple instances and address families
	return 0
}

type interfaceManager struct {
	intf   *net.Interface
	conn   *ipv6.PacketConn
	area   *areaManager
	config InterfaceConfig

	state     InterfaceState
	neighbors map[uint32]*neighbor
	dr        packet.ID
	bdr       packet.ID

	events   chan InterfaceEvent
	messages chan messageWrapper

	log logrus.FieldLogger
}

const ipNetwork = "ip6:89"
const addrAllSPFRouters = "FF02::5"
const addrAllDRouters = "FF02::6"

func findLinkLocalAddress(intf *net.Interface) (*net.IP, error) {
	addrs, err := intf.Addrs()
	if err != nil {
		return nil, errors.Wrap(err, "unable to enumerate interface addresses")
	}

	for _, ifAddr := range addrs {
		addr := net.ParseIP(ifAddr.String())
		if addr.IsLinkLocalUnicast() {
			return &addr, nil
		}
	}

	return nil, errors.New("unable to find link local address")
}

func makeConn(intf *net.Interface) (*ipv6.PacketConn, error) {
	addr, err := findLinkLocalAddress(intf)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find address for interface")
	}

	llAddr := fmt.Sprintf("%s%%%s", addr.String(), intf.Name)
	pc, err := net.ListenPacket(ipNetwork, llAddr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to listen on unicast address")
	}

	allSPFRouters := net.IPAddr{IP: net.ParseIP("ff02::5")}
	conn := ipv6.NewPacketConn(pc)
	if err := conn.JoinGroup(intf, &allSPFRouters); err != nil {
		return nil, errors.Wrap(err, "unable to join AllSPFRouters multicast group")
	}

	return conn, nil
}

func newInterfaceManager(log logrus.FieldLogger, area *areaManager, name string, config InterfaceConfig) (*interfaceManager, error) {
	intf, err := net.InterfaceByName(name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find interface")
	}

	if config.IfType == 0 {
		guessType, err := guessIfType(intf)
		if err != nil {
			return nil, errors.Wrap(err, "interface type unset, unable to guess")
		}
		config.IfType = guessType
	}

	mgmt := &interfaceManager{
		intf:     intf,
		area:     area,
		config:   config.ApplyDefaults(),
		events:   make(chan InterfaceEvent),
		messages: make(chan messageWrapper),
		log:      log.WithField("interface", name),
	}

	return mgmt, nil
}

func (im *interfaceManager) Start(ctx context.Context) error {
	conn, err := makeConn(im.intf)
	if err != nil {
		return errors.Wrap(err, "unable to create IP listener")
	}
	im.conn = conn

	go im.processIncomingMessages(ctx)
	go im.execute(ctx)
	go func(ctx context.Context) {
		<-ctx.Done()
		im.cleanup()
	}(ctx)

	return nil
}

func (im *interfaceManager) SetLinkUp() {
	im.events <- IfEventInterfaceUp
}

func (im *interfaceManager) SetLinkDown() {
	im.events <- IfEventInterfaceDown
}

func (im *interfaceManager) GetInterface() *net.Interface {
	return im.intf
}

func (im *interfaceManager) GetConfig() InterfaceConfig {
	return im.config
}

func (im *interfaceManager) IsDR() bool {
	return im.area.GetConfig().routerID == im.dr
}

func (im *interfaceManager) IsBDR() bool {
	return im.area.GetConfig().routerID == im.bdr
}

func (im *interfaceManager) runEvent(event InterfaceEvent) {
	im.events <- event
}

func (im *interfaceManager) cleanup() {
	im.conn.Close()
}

type messageWrapper struct {
	src net.IP
	msg *packet.OSPFv3Message
}

func (im *interfaceManager) processIncomingMessages(ctx context.Context) {
	buf := new(bytes.Buffer)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		bytes := make([]byte, 1500)
		_, _, addr, err := im.conn.ReadFrom(bytes)
		if err != nil {
			im.log.WithError(err).Error("Reading packet failed")
			continue
		}
		buf.Reset()
		buf.Write(bytes)

		msg, _, err := packet.DeserializeOSPFv3Message(buf)
		if err != nil {
			im.log.WithError(err).Error("Failed deserializing message")
			continue
		}

		src := net.ParseIP(addr.String())
		pl := messageWrapper{
			msg: msg,
			src: src,
		}

		im.messages <- pl
	}
}

func (im *interfaceManager) execute(ctx context.Context) {
	helloTimer := time.NewTicker(im.config.HelloInterval)
	rtmxTimer := time.NewTicker(im.config.RxmtInterval)
	defer func() {
		helloTimer.Stop()
		rtmxTimer.Stop()
	}()

	for {
		var err error
		select {
		case <-ctx.Done():
			return
		case msg, open := <-im.messages:
			if !open {
				return
			}
			err = im.handleMessage(ctx, msg)
		case event, open := <-im.events:
			if !open {
				return
			}
			err = im.handleEvent(event)
		case t := <-helloTimer.C:
			err = im.sendHello(t, nil)
		case t := <-rtmxTimer.C:
			err = im.retransmit(t)
		}

		if err != nil {
			im.log.WithError(err).Error("Execution error")
		}
	}
}

func (im *interfaceManager) handleMessage(ctx context.Context, pl messageWrapper) error {
	var err error
	switch body := pl.msg.Body.(type) {
	case *packet.Hello:
		im.handleHello(ctx, pl.src, pl.msg, body)
	}

	return err
}

// runs in sync with state machine
func (im *interfaceManager) handleHello(ctx context.Context, src net.IP, header *packet.OSPFv3Message, pl *packet.Hello) {
	eBit := (pl.Options.Flags & packet.RouterOptE) != 0
	if pl.HelloInterval != uint16(im.config.HelloInterval.Seconds()) ||
		pl.RouterDeadInterval != uint16(im.config.RouterDeadInterval.Seconds()) ||
		header.AreaID != im.area.GetConfig().ID ||
		eBit != !im.area.GetConfig().Stub {
		im.log.
			WithFields(logrus.Fields{
				"src":                src,
				"helloInterval":      pl.HelloInterval,
				"routerDeadInterval": pl.RouterDeadInterval,
				"area":               header.AreaID,
				"externalRoutingCap": eBit,
			}).
			Debug("discarding hello packet for mismatch in confguration")
		return
	}

	neighborID := uint32(header.RouterID)
	// todo: locking
	neigh, found := im.neighbors[neighborID]
	if !found {
		neigh = newNeighbor(im, header.RouterID, pl.RouterPriority)
		neigh.start(ctx)
		im.neighbors[neighborID] = neigh
	}

	prevAttrs := neigh.SetAttributes(NeighborAttributes{
		priority: pl.RouterPriority,
		address:  src,
		options:  pl.Options,
		dr:       pl.DesignatedRouterID,
		bdr:      pl.BackupDesignatedRouterID,
	})

	neighWasDR := prevAttrs.dr == neigh.id
	neighWasBDR := prevAttrs.bdr == neigh.id

	neigh.RunEvent(NbrEventHelloReceived)

	for _, n := range pl.Neighbors {
		if n == im.area.GetConfig().routerID {
			neigh.RunEvent(NbrEvent2WayReceived)
			break
		}
	}

	if prevAttrs.priority != pl.RouterPriority {
		im.runEvent(IfEventNeighborChange)
	}

	if pl.DesignatedRouterID == neigh.id &&
		pl.BackupDesignatedRouterID == 0 &&
		im.state == IfStateWaiting {
		im.runEvent(IfEventBackupSeen)
	} else if neighWasDR != (pl.DesignatedRouterID == neigh.id) {
		im.runEvent(IfEventNeighborChange)
	}

	if pl.BackupDesignatedRouterID == neigh.id && im.state == IfStateWaiting {
		im.runEvent(IfEventBackupSeen)
	} else if neighWasBDR != (pl.BackupDesignatedRouterID == neigh.id) {
		im.runEvent(IfEventNeighborChange)
	}
}

func (im *interfaceManager) handleEvent(event InterfaceEvent) error {
	// TODO: Interface State Machine (https://tools.ietf.org/html/rfc2328#page-72)
	return nil
}

// needs to be goroutine safe! (config is expected to be read-only)
func (im *interfaceManager) sendMessage(dest net.Addr, msgType packet.OSPFMessageType, body packet.Serializable) error {
	msg := packet.OSPFv3Message{
		Version:    ospfVersion,
		Type:       msgType,
		RouterID:   im.area.GetConfig().routerID,
		AreaID:     im.area.GetConfig().ID,
		InstanceID: im.config.getInstanceID(),
		Body:       body,
	}

	buf := bytes.NewBuffer(nil)
	msg.Serialize(buf)

	_, err := im.conn.WriteTo(buf.Bytes(), nil, dest)
	return err
}

func (im *interfaceManager) sendHello(t time.Time, nbr *neighbor) error {
	// TODO
	return nil
}

func (im *interfaceManager) sendDatabaseDescription(nbr *neighbor, pl packet.DatabaseDescription) error {
	// TODO
	return nil
}

func (im *interfaceManager) retransmit(t time.Time) error {
	// TODO
	return nil
}
