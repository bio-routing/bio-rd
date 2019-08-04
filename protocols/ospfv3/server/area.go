package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/bio-routing/bio-rd/protocols/ospfv3/packet"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type AreaConfig struct {
	ID              packet.ID
	routerID        packet.ID
	Stub            bool // in spec ExternalRoutingCapability (changed so default is false)
	StubDefaultCost packet.InterfaceMetric
}

type areaManager struct {
	interfaces      map[string]*interfaceManager
	interfacesMutex sync.Mutex

	config  AreaConfig
	started *uint64

	routerLSAs  []packet.RouterLSA
	networkLSAs []packet.NetworkLSA
	lsaMutex    sync.Mutex

	tree      spfTree
	treeMutex sync.Mutex

	rib RoutingTable
	log logrus.FieldLogger
}

func newAreaManager(log logrus.FieldLogger, config AreaConfig, rib RoutingTable) (*areaManager, error) {
	mgmt := &areaManager{
		interfaces: make(map[string]*interfaceManager),

		config: config,
		rib:    rib,
		log:    log.WithField("area", fmt.Sprintf("%d", config.ID)),
	}

	return mgmt, nil
}

func (am *areaManager) AddInterface(ctx context.Context, name string, config InterfaceConfig) error {
	am.interfacesMutex.Lock()
	defer am.interfacesMutex.Unlock()

	if _, ok := am.interfaces[name]; ok {
		return errors.New("interface is already configured. cannot reconfigure")
	}

	ifm, err := newInterfaceManager(am.log.WithField("component", "interfaceManager"), am, name, config)
	if err != nil {
		return errors.Wrap(err, "unable to construct interface manager")
	}

	am.interfaces[name] = ifm
	if am.Started() {
		if err := ifm.Start(ctx); err != nil {
			return errors.Wrap(err, "unable to start interface manager")
		}
	}

	return nil
}

func (am *areaManager) Started() bool {
	return atomic.LoadUint64(am.started) == 1
}

func (am *areaManager) Start(ctx context.Context) error {
	if atomic.SwapUint64(am.started, 1) == 1 {
		return errors.New("already started")
	}

	for name, ifm := range am.interfaces {
		if err := ifm.Start(ctx); err != nil {
			return errors.Wrap(err, fmt.Sprintf("unable to start interface: %s", name))
		}
		intf := ifm.GetInterface()
		if intf != nil && (intf.Flags&net.FlagUp) != 0 {
			ifm.SetLinkUp()
		}
	}

	if err := am.startMonitorLinks(ctx); err != nil {
		return errors.Wrap(err, "unable to listen for interface link changes")
	}

	return nil
}

func (am *areaManager) GetConfig() AreaConfig {
	return am.config
}

func (am *areaManager) startMonitorLinks(ctx context.Context) error {
	updateCh := make(chan netlink.LinkUpdate)
	done := make(chan struct{})
	if err := netlink.LinkSubscribe(updateCh, done); err != nil {
		return err
	}

	go func() {
		for {
			var update netlink.LinkUpdate
			select {
			case <-ctx.Done():
				close(done)
				return
			case update = <-updateCh:
			}

			name := update.Link.Attrs().Name
			am.interfacesMutex.Lock()
			ifm, found := am.interfaces[name]
			am.interfacesMutex.Unlock()
			if !found {
				continue
			}

			if (net.Flags(update.IfInfomsg.Flags) & net.FlagUp) != 0 {
				ifm.SetLinkUp()
			} else {
				ifm.SetLinkDown()
			}
		}
	}()

	return nil
}
