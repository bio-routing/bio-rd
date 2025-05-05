package vamp

import (
	"fmt"
	"sync"
	"time"

	vapi "github.com/bio-routing/bio-rd/protocols/vamp/api"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
)

const (
	reconnectInterval = time.Second * 10
)

type ClientConfig struct {
	ServerAddr  string
	VRFRegistry *vrf.VRFRegistry
	VRFs        []string
}

type VAMPClient struct {
	serverAddr  string
	vrfRegistry *vrf.VRFRegistry
	vrfs        []*vrf.VRF
	vrfTracker  *vrfTracker

	clientMu     sync.Mutex // Guards ribOberservs + updateCh
	ribOberservs []*RIBObserver
	updateCh     chan *vapi.VAMPMessage
}

func NewClient(cfg ClientConfig) (*VAMPClient, error) {
	vc := &VAMPClient{
		serverAddr:  cfg.ServerAddr,
		vrfRegistry: cfg.VRFRegistry,
		vrfs:        make([]*vrf.VRF, len(cfg.VRFs)),
		vrfTracker:  newVRFTracker(),
		updateCh:    make(chan *vapi.VAMPMessage),
	}

	if cfg.VRFRegistry == nil {
		return nil, fmt.Errorf("VRFRegistry config attribute must be set")
	}

	if len(cfg.VRFs) == 0 {
		return nil, fmt.Errorf("no VRFs specified in configuration")
	}

	for i, vrfName := range cfg.VRFs {
		vrf := vc.vrfRegistry.GetVRFByName(vrfName)
		if vrf == nil {
			return nil, fmt.Errorf("unknown VRF %q", vrfName)
		}

		_, err := vc.vrfTracker.registerVRF(vrfName)
		if err != nil {
			return nil, err
		}

		vc.vrfs[i] = vrf
	}

	return vc, nil
}

func (vc *VAMPClient) Init() error {
	return vc.setupPlumbing()
}

func (vc *VAMPClient) Start() {
	vc.clientMu.Lock()
	defer vc.clientMu.Unlock()

	for _, ribObserver := range vc.ribOberservs {
		go ribObserver.Start()
	}
}

func (vc *VAMPClient) Stop() {
	vc.clientMu.Lock()
	defer vc.clientMu.Unlock()

	for _, ribObserver := range vc.ribOberservs {
		go ribObserver.Stop()
	}
}

func (vc *VAMPClient) setupPlumbing() error {
	err := vc.connect()
	if err != nil {
		return err
	}

	return vc.setupRIBObservers()
}

func (vc *VAMPClient) connect() error {
	// TODO: Set up gRPC connection to vc.serverAddr
	return nil
}

func (vc *VAMPClient) setupRIBObservers() error {
	vc.clientMu.Lock()
	defer vc.clientMu.Unlock()

	if len(vc.ribOberservs) != 0 {
		return fmt.Errorf("RIBObservers seem to have already been set up")
	}

	// (Re)create update channel (HERE?)
	vc.updateCh = make(chan *vapi.VAMPMessage)

	ribOberservs := make([]*RIBObserver, 0)
	for _, vrf := range vc.vrfs {
		vrfID, err := vc.vrfTracker.getVRFID(vrf.Name())
		if err != nil {
			return fmt.Errorf("failed to get VRF ID for VRF %q: %v", vrf.Name(), err)
		}

		ipv4RIB := vrf.IPv4UnicastRIB()
		if ipv4RIB != nil {
			ribOberservs = append(ribOberservs, newRIBOserver(ipv4RIB, vrfID, vc.updateCh))
		}

		ip64RIB := vrf.IPv6UnicastRIB()
		if ipv4RIB != nil {
			ribOberservs = append(ribOberservs, newRIBOserver(ip64RIB, vrfID, vc.updateCh))
		}
	}

	vc.ribOberservs = ribOberservs

	return nil
}
