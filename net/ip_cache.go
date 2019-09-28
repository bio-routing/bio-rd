package net

import (
	"sync"
)

const (
	ipCacheInitialSize = 1000000
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
		cache: make(map[IP]*IP, ipCacheInitialSize),
	}
}

func (ipc *ipCache) get(addr IP) *IP {
	ipc.cacheMu.Lock()

	if x, ok := ipc.cache[addr]; ok {
		ipc.cacheMu.Unlock()
		return x
	}

	newAddr := addr.Copy()
	ipc.cache[addr] = newAddr
	ipc.cacheMu.Unlock()

	return newAddr
}
