package net

import (
	"sync"
)

const (
	ipCachePreAlloc = 0
)

var (
	ipc *ipCache
)

func init() {
	ipc = newIPCache()
}

type ipCache struct {
	cache   map[IP]*IP
	cacheMu sync.Mutex
}

func newIPCache() *ipCache {
	return &ipCache{
		cache: make(map[IP]*IP, ipCachePreAlloc),
	}
}

func (ipc *ipCache) get(addr IP) *IP {
	ipc.cacheMu.Lock()

	if a, exists := ipc.cache[addr]; exists {
		ipc.cacheMu.Unlock()
		return a
	}

	ipc._set(addr)
	res := ipc.cache[addr]
	ipc.cacheMu.Unlock()
	return res
}

func (ipc *ipCache) _set(addr IP) {
	ipc.cache[addr] = &addr
}
