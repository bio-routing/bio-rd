package server

import (
	"fmt"
	"sync"

	"github.com/bio-routing/bio-rd/util/log"
)

type netIfaManager struct {
	srv           *Server
	netIfas       map[string]*netIfa
	netIfasMu     sync.Mutex
	useMockTicker bool
}

func newNetIfaManager(srv *Server) *netIfaManager {
	return &netIfaManager{
		srv:     srv,
		netIfas: make(map[string]*netIfa),
	}
}

// AddInterface adds an interface to the ISIS server
func (s *Server) AddInterface(cfg *InterfaceConfig) error {
	log.Debugf("IS-IS: Adding interface %s", cfg.Name)
	return s.netIfaManager.addInterface(cfg)
}

func (s *Server) RemoveInterface(name string) error {
	log.Debugf("IS-IS: Removing interface %s", name)
	return s.netIfaManager.removeInterface(name)
}

func (nima *netIfaManager) removeInterface(name string) error {
	nima.netIfasMu.Lock()
	defer nima.netIfasMu.Unlock()

	if _, exists := nima.netIfas[name]; !exists {
		return fmt.Errorf("IS-IS is not enabled on interface %q", name)
	}

	nima.netIfas[name].stop()
	delete(nima.netIfas, name)

	return nil
}

func (nima *netIfaManager) addInterface(cfg *InterfaceConfig) error {
	nima.netIfasMu.Lock()
	defer nima.netIfasMu.Unlock()

	if _, exists := nima.netIfas[cfg.Name]; exists {
		return fmt.Errorf("ISIS is enabled on that interface already. Updating config is not supported yet")
	}

	ifa := newNetIfa(nima.srv, cfg)
	nima.netIfas[cfg.Name] = ifa

	return nil
}

func (nima *netIfaManager) getInterface(name string) *netIfa {
	nima.netIfasMu.Lock()
	defer nima.netIfasMu.Unlock()

	if _, found := nima.netIfas[name]; !found {
		return nil
	}

	return nima.netIfas[name]
}

func (nima *netIfaManager) getAllInterfacesExcept(exception *netIfa) []*netIfa {
	nima.netIfasMu.Lock()
	defer nima.netIfasMu.Unlock()

	res := make([]*netIfa, 0)
	for _, ifa := range nima.netIfas {
		if ifa != exception {
			res = append(res, ifa)
		}
	}

	return res
}

func (nima *netIfaManager) getAllInterfaces() []*netIfa {
	nima.netIfasMu.Lock()
	defer nima.netIfasMu.Unlock()

	res := make([]*netIfa, 0)
	for _, ifa := range nima.netIfas {
		res = append(res, ifa)
	}

	return res
}
