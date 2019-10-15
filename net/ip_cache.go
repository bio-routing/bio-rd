package net

import (
	"sync"

	"github.com/google/btree"
)

const (
	ipCacheBTreeGrade = 3500
)

var (
	ipc *ipCache
)

func init() {
	ipc = newIPCache()
}

type ipCache struct {
	cacheMu sync.Mutex
	tree    *btree.BTree
}

func newIPCache() *ipCache {
	return &ipCache{
		tree: btree.New(ipCacheBTreeGrade),
	}
}

func (ipc *ipCache) get(addr *IP) *IP {
	ipc.cacheMu.Lock()

	item := ipc.tree.Get(addr)
	if item != nil {
		ipc.cacheMu.Unlock()

		return item.(*IP)
	}

	ipc.tree.ReplaceOrInsert(addr)
	ipc.cacheMu.Unlock()

	return addr
}
