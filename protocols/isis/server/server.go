package server

import (
	"fmt"
	"sync"
	"syscall"
	"time"

	"github.com/bio-routing/bio-rd/config"
	"github.com/bio-routing/bio-rd/protocols/isis/packet"
	"github.com/bio-routing/bio-rd/protocols/isis/types"
	log "github.com/sirupsen/logrus"
)

const (
	maxEtherFrameSize = 9216
)

//ISISServer represents an ISIS speaker
type ISISServer struct {
	config         config.ISISConfig
	sequenceNumber uint32
	interfaces     map[string]*netIf
	interfacesMu   sync.RWMutex
	lsdb           *lsdb
	stop           chan struct{}
}

type isisNeighbor struct {
	SystemID  types.SystemID
	Interface string
}

// NewISISServer creates and initializes a new ISIS speaker
func NewISISServer(cfg *config.ISISConfig) *ISISServer {
	server := &ISISServer{
		config:         *cfg,
		sequenceNumber: 1,
		interfaces:     make(map[string]*netIf),
		stop:           make(chan struct{}),
	}

	server.lsdb = newLSDB(server)
	return server
}

// Start starts an ISIS speaker
func (isis *ISISServer) Start() error {
	go func() {
		t := time.NewTicker(time.Second)
		for {
			select {
			case <-isis.stop:
				return
			case <-t.C:
				isis.lsdb.decrementRemainingLifetimes()
			}
		}
	}()

	for _, ifa := range isis.config.Interfaces {
		err := isis.AddInterface(ifa)
		if err != nil {
			log.Errorf("Failed to activste ISIS on interface %s: %v", ifa.Name, err)
		}
	}

	return nil
}

// AddInterface adds a network interface to the ISIS server
func (isis *ISISServer) AddInterface(ifa config.ISISInterfaceConfig) error {
	isis.interfacesMu.Lock()
	defer isis.interfacesMu.Unlock()

	if _, ok := isis.interfaces[ifa.Name]; ok {
		return fmt.Errorf("Interface exists already")
	}

	interf, err := newNetIf(isis, ifa)
	if err != nil {
		return fmt.Errorf("Unable to enable ISIS on %s: %v", ifa.Name, err)
	}

	isis.interfaces[ifa.Name] = interf
	isis.interfaces[ifa.Name].startReceiver()
	isis.interfaces[ifa.Name].helloSender()

	return nil
}

// RemoveInterface removes an interface from the ISIS server
func (isis *ISISServer) RemoveInterface(ifName string) error {
	isis.interfacesMu.Lock()
	defer isis.interfacesMu.Unlock()

	if _, ok := isis.interfaces[ifName]; !ok {
		return fmt.Errorf("Interface does not exist")
	}

	isis.stopInterface(isis.interfaces[ifName])
	return nil
}

// Stop stops an ISIS speaker
func (isis *ISISServer) Stop() error {
	for _, ifa := range isis.interfaces {
		isis.stopInterface(ifa)
	}

	isis.stop <- struct{}{}
	return nil
}

func (isis *ISISServer) stopInterface(ifa *netIf) {
	// stop the hello sender and frame receiver
	ifa.stop <- struct{}{}
	ifa.stop <- struct{}{}

	syscall.Close(ifa.socket)

	// TODO: Neighbor tear down

	delete(isis.interfaces, ifa.name)
	isis.lsdb.clearSRMSSN(ifa) // Possible race condition: How to we make sure received LSPs are not maked with SRM for this interface?
	ifa.isisServer = nil
}

func (isis *ISISServer) systemID() [6]byte {
	return isis.config.NETs[0].SystemID
}

func (isis *ISISServer) lsp() *packet.LSPDU {
	lspdu := &packet.LSPDU{
		RemainingLifetime: 1200,
		LSPID: packet.LSPID{
			SystemID:     isis.systemID(),
			PseudonodeID: 0,
			/*SequenceNumber: isis.sequenceNumber,
			TypeBlock:      3,
			TLVs:           make(packet.TLV, 0),*/
		},
	}

	// TODO: TLVs

	return lspdu
}

func (isis *ISISServer) DumpLSDB() {
	isis.lsdb.lspsMu.RLock()
	defer isis.lsdb.lspsMu.RUnlock()

	fmt.Printf("LSDB Dump:\n")

	for id, e := range isis.lsdb.lsps {
		fmt.Printf("ID: %s.%d\n", id.SystemID.String(), id.PseudonodeID)
		fmt.Printf("Checksum: 0x%x\n", e.lspdu.Checksum)
		fmt.Printf("Length: %d\n", e.lspdu.Length)
		fmt.Printf("Remaining Lifetime: %d\n", e.lspdu.RemainingLifetime)
		fmt.Printf("Sequence Number: %d\n", e.lspdu.SequenceNumber)
		fmt.Printf("Type Block: %d\n", e.lspdu.TypeBlock)
	}
}
