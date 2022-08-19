package route

import (
	"sync"

	"github.com/bio-routing/bio-rd/util/log"
)

// BGPPathManager is a component used to deduplicate BGP Path objects
type BGPPathManager struct {
	paths map[string]*BGPPathCounter
	mu    sync.Mutex
}

// BGPPathCounter couples a counter to a particular path
type BGPPathCounter struct {
	usageCount uint64
	path       *BGPPath
}

// NewBGPPathManager creates a new BGP Path Manager
func NewBGPPathManager() *BGPPathManager {
	m := &BGPPathManager{}
	return m
}

func (m *BGPPathManager) lookup(p BGPPath) *BGPPath {
	pathCounter, ok := m.paths[p.ComputeHash()]
	if !ok {
		return nil
	}

	return pathCounter.path
}

// AddPath adds a path to the cache if it doesn't exist. If it exist a pointer to the cached object is returned.
func (m *BGPPathManager) AddPath(p BGPPath) *BGPPath {
	m.mu.Lock()
	defer m.mu.Unlock()

	hash := p.ComputeHash()

	q := m.lookup(p)
	if q == nil {
		m.paths[hash] = &BGPPathCounter{
			path: &p,
		}
	}

	m.paths[hash].usageCount++
	return m.paths[hash].path
}

// RemovePath notifies us that there is one user less for path p
func (m *BGPPathManager) RemovePath(p BGPPath) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.lookup(p) == nil {
		log.Errorf("Tried to remove non-existent BGPPath: %v", p)
		return
	}

	hash := p.ComputeHash()

	m.paths[hash].usageCount--
	if m.paths[hash].usageCount == 0 {
		delete(m.paths, hash)
	}
}
