package route

import "sync"

const initialBGPPathACacheSize = 100000

var (
	bgpC *bgpPathACache
)

func init() {
	bgpC = newBGPPathACache()
}

type bgpPathACache struct {
	cache   map[BGPPathA]*BGPPathA
	cacheMu sync.Mutex
}

func newBGPPathACache() *bgpPathACache {
	return &bgpPathACache{
		cache: make(map[BGPPathA]*BGPPathA, initialBGPPathACacheSize),
	}
}

func (bgpc *bgpPathACache) get(p *BGPPathA) *BGPPathA {
	bgpc.cacheMu.Lock()

	if x, ok := bgpc.cache[*p]; ok {
		bgpc.cacheMu.Unlock()
		return x
	}

	bgpc.cache[*p] = p
	bgpc.cacheMu.Unlock()
	return p
}
