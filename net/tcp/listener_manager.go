package tcp

import (
	"fmt"
	"net"
	"sync"

	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/bio-routing/bio-rd/util/log"
)

type ConnWithVRF struct {
	Conn net.Conn
	VRF  *vrf.VRF
}

type ListenerManagerI interface {
	ListenAddrsPerVRF(vrf *vrf.VRF) []string
	GetListeners(v *vrf.VRF) []ListenerI
	CreateListenersIfNotExists(v *vrf.VRF) error
	AcceptCh() chan ConnWithVRF
}

type ListenerManager struct {
	listenAddrsByVRF map[string][]string
	listenersByVRF   map[string][]ListenerI
	listenersByVRFmu sync.RWMutex
	acceptCh         chan ConnWithVRF
	listenerFactory  ListenerFactoryI
}

func NewListenerManager(listenAddrsByVRF map[string][]string) *ListenerManager {
	return &ListenerManager{
		listenAddrsByVRF: listenAddrsByVRF,
		listenersByVRF:   make(map[string][]ListenerI),
		listenerFactory:  NewListenerFactory(),
		acceptCh:         make(chan ConnWithVRF),
	}
}

func (lm *ListenerManager) SetListenerFactory(lf ListenerFactoryI) {
	lm.listenerFactory = lf
}

func (lm *ListenerManager) ListenAddrsPerVRF(vrf *vrf.VRF) []string {
	return lm.listenAddrsByVRF[vrf.Name()]
}

func (lm *ListenerManager) GetListeners(v *vrf.VRF) []ListenerI {
	lm.listenersByVRFmu.Lock()
	defer lm.listenersByVRFmu.Unlock()

	ret := make([]ListenerI, 0)
	for _, l := range lm.listenersByVRF[v.Name()] {
		ret = append(ret, l)
	}

	return ret
}

func (lm *ListenerManager) CreateListenersIfNotExists(v *vrf.VRF) error {
	lm.listenersByVRFmu.Lock()
	defer lm.listenersByVRFmu.Unlock()

	if _, exists := lm.listenersByVRF[v.Name()]; exists {
		return nil
	}

	err := lm._createListeners(v)
	if err != nil {
		return fmt.Errorf("unable to create listeners: %v", err)
	}

	return nil
}

func (lm *ListenerManager) _createListeners(v *vrf.VRF) error {
	for _, addr := range lm.ListenAddrsPerVRF(v) {
		err := lm._addListener(v, addr, lm.acceptCh)
		if err != nil {
			return fmt.Errorf("unable to create TCP listener %q vrf %s: %v", addr, v.Name(), err)
		}
	}

	return nil
}

// newListener creates a new Listener
func (lm *ListenerManager) _addListener(vrf *vrf.VRF, addr string, ch chan ConnWithVRF) error {
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	log.Infof("Listener manager: Starting TCP listener on %s in VRF %s", addr, vrf.Name())
	l, err := lm.listenerFactory.NewListener(vrf, tcpaddr, 255)
	if err != nil {
		return err
	}

	lm._add(vrf, l)

	go func(tl ListenerI) error {
		defer lm.dropListener(vrf, tl)

		for {
			conn, err := l.AcceptTCP()
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"Topic": "Peer",
				}).Error("Failed to AcceptTCP")
				return err
			}

			ch <- ConnWithVRF{
				Conn: conn,
				VRF:  vrf,
			}
		}
	}(l)

	return nil
}

// _add is to be called with the mutex acquired
func (lm *ListenerManager) _add(vrf *vrf.VRF, l ListenerI) {
	if _, exists := lm.listenersByVRF[vrf.Name()]; !exists {
		lm.listenersByVRF[vrf.Name()] = make([]ListenerI, 0)
	}

	lm.listenersByVRF[vrf.Name()] = append(lm.listenersByVRF[vrf.Name()], l)
}

func (lm *ListenerManager) dropListener(vrf *vrf.VRF, l ListenerI) {
	lm.listenersByVRFmu.Lock()
	defer lm.listenersByVRFmu.Unlock()

	vrfName := vrf.Name()
	listeners := lm.listenersByVRF[vrfName]
	for i, x := range listeners {
		if x == l {
			lm.listenersByVRF[vrfName] = append(listeners[:i], listeners[i+1:]...)
			return
		}
	}
}

func (lm *ListenerManager) AcceptCh() chan ConnWithVRF {
	return lm.acceptCh
}
