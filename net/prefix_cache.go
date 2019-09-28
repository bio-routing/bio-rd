package net

import "sync"

const (
	prefixCacheInitialSize = 1000000
)

var (
	pfxc *pfxCache
)

func init() {
	pfxc = newPfxCache()
}

type pfxCache struct {
	cache   map[Prefix]*Prefix
	cacheMu sync.Mutex
}

func newPfxCache() *pfxCache {
	return &pfxCache{
		cache: make(map[Prefix]*Prefix, prefixCacheInitialSize),
	}
}

func (pfxc *pfxCache) get(pfx Prefix) *Prefix {
	pfxc.cacheMu.Lock()

	if x, ok := pfxc.cache[pfx]; ok {
		pfxc.cacheMu.Unlock()
		return x
	}

	c := pfx.Copy()
	pfxc.cache[pfx] = c
	pfxc.cacheMu.Unlock()

	return c
}
