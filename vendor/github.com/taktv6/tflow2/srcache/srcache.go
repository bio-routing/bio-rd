package srcache

import (
	"net"
	"sync"

	"github.com/taktv6/tflow2/config"
)

// SamplerateCache caches information about samplerates
type SamplerateCache struct {
	cache map[string]uint64
	mu    sync.RWMutex
}

// New creates a new SamplerateCache and initializes it with values from the config
func New(agents []config.Agent) *SamplerateCache {
	c := &SamplerateCache{
		cache: make(map[string]uint64),
	}

	// Initialize cache with configured samplerates
	for _, a := range agents {
		c.Set(net.ParseIP(*a.IPAddress), *a.SampleRate)
	}

	return c
}

// Set updates a cache entry
func (s *SamplerateCache) Set(rtr net.IP, rate uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[string(rtr)] = rate
}

// Get gets a cache entry
func (s *SamplerateCache) Get(rtr net.IP) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.cache[string(rtr)]; !ok {
		return 1
	}

	return s.cache[string(rtr)]
}
