package locRIB

import (
	"fmt"
	"sync"
)

var defaultRegistry *registry

func init() {
	defaultRegistry = &registry{
		ribs: make(map[string]*LocRIB),
	}
}

type registry struct {
	ribs map[string]*LocRIB
	mu   sync.Mutex
}

func LocRIBByName(name string) *LocRIB {
	rib, _ := defaultRegistry.ribs[name]
	return rib
}

func (r *registry) register(rib *LocRIB) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, found := r.ribs[rib.name]; found {
		return fmt.Errorf(fmt.Sprintf("a rib with name '%s' already exists", rib.name))
	}

	r.ribs[rib.name] = rib

	return nil
}
