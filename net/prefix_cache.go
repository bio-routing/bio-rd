package net

import (
	"sync"

	"github.com/google/btree"
)

const (
	prefixCacheBTreeGrade = 3500
)

var (
	pfxc *pfxCache
)

func init() {
	pfxc = newPfxCache()
}

type pfxCache struct {
	cacheMu sync.Mutex
	tree    *btree.BTree
}

func newPfxCache() *pfxCache {
	return &pfxCache{
		tree: btree.New(3500),
	}
}

func (pfxc *pfxCache) get(pfx *Prefix) *Prefix {
	pfxc.cacheMu.Lock()

	item := pfxc.tree.Get(pfx)
	if item != nil {
		pfxc.cacheMu.Unlock()
		return item.(*Prefix)
	}

	pfxc.tree.ReplaceOrInsert(pfx)
	pfxc.cacheMu.Unlock()

	return pfx
}
