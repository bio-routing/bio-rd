package rt

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type BGPPath struct {
	PathIdentifier uint32
	NextHop        uint32
	LocalPref      uint32
	ASPath         string
	ASPathLen      uint16
	Origin         uint8
	MED            uint32
	EBGP           bool
	Source         uint32
}

type BGPPathManager struct {
	paths map[BGPPath]*BGPPathCounter
	mu    sync.Mutex
}

type BGPPathCounter struct {
	usageCount uint64
	path       *BGPPath
}

func NewBGPPathManager() *BGPPathManager {
	m := &BGPPathManager{}
	return m
}

func (m *BGPPathManager) pathExists(p BGPPath) bool {
	if _, ok := m.paths[p]; !ok {
		return false
	}

	return true
}

func (m *BGPPathManager) AddPath(p BGPPath) *BGPPath {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.pathExists(p) {
		m.paths[p] = &BGPPathCounter{
			path: &p,
		}
	}

	m.paths[p].usageCount++
	return m.paths[p].path
}

func (m *BGPPathManager) RemovePath(p BGPPath) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.pathExists(p) {
		log.Fatalf("Tried to remove non-existent BGPPath: %v", p)
		return
	}

	m.paths[p].usageCount--
	if m.paths[p].usageCount == 0 {
		delete(m.paths, p)
	}
}

func (r *Route) bgpPathSelection() (res []*Path) {
	// TODO: Implement next hop lookup and compare IGP metrics
	if len(r.paths) == 1 {
		copy(res, r.paths)
		return res
	}

	for _, p := range r.paths {
		if p.Type != BGPPathType {
			continue
		}

		if len(res) == 0 {
			res = append(res, p)
			continue
		}

		if res[0].BGPPath.ecmp(p.BGPPath) {
			res = append(res, p)
			continue
		}

		if !res[0].BGPPath.better(p.BGPPath) {
			continue
		}

		res = []*Path{p}
	}

	return res
}

func (b *BGPPath) better(c *BGPPath) bool {
	if c.LocalPref < b.LocalPref {
		return false
	}

	if c.LocalPref > b.LocalPref {
		return true
	}

	if c.ASPathLen > b.ASPathLen {
		return false
	}

	if c.ASPathLen < b.ASPathLen {
		return true
	}

	if c.Origin > b.Origin {
		return false
	}

	if c.Origin < b.Origin {
		return true
	}

	if c.MED > b.MED {
		return false
	}

	if c.MED < b.MED {
		return true
	}

	return false
}

func (b *BGPPath) ecmp(c *BGPPath) bool {
	return b.LocalPref == c.LocalPref && b.ASPathLen == c.ASPathLen && b.Origin == c.Origin && b.MED == c.MED
}
