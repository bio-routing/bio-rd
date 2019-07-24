package net

import "sync"

const (
	prefixCacheInitialSize = 1000000
)

var (
	pfxc *pfxCache
)

func init() {
	ipc = newIPCache()
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

func (pfxc *pfxCache) get(pfx *Prefix) *Prefix {
	pfxc.cacheMu.Lock()

	if _, ok := pfxc.cache[*pfx]; ok {
		x := pfxc.cache[*pfx]
		pfxc.cacheMu.Unlock()
		return x
	}

	pfxc.cache[*pfx] = pfx
	pfxc.cacheMu.Unlock()

	return pfx
}
