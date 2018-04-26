package database

import (
	"fmt"
	"net"
	"sync"

	"github.com/taktv6/tflow2/avltree"
	"github.com/taktv6/tflow2/convert"
)

type mapTree struct {
	entries map[string]*avltree.Tree
	sync.RWMutex
}

func newMapTree() *mapTree {
	return &mapTree{
		entries: make(map[string]*avltree.Tree),
	}
}

func createKey(key interface{}) string {
	switch val := key.(type) {
	case string:
		return val
	case []uint8:
		return string(val)
	case byte:
		return string([]byte{val})
	case int64:
		return string(convert.Int64Byte(val))
	case uint16:
		return string(convert.Uint16Byte(val))
	case uint32:
		return string(convert.Uint32Byte(val))
	case net.IP:
		if addr := val.To4(); addr != nil {
			return string(addr)
		}
		return string(val.To16())
	default:
		panic(fmt.Sprintf("unsupported key type: %T", key))
	}
}

func (m *mapTree) Insert(key interface{}, value interface{}) {
	keyStr := createKey(key)
	m.Lock()
	root, ok := m.entries[keyStr]
	if !ok {
		root = avltree.New()
		m.entries[keyStr] = root
	}
	root.Insert(value, value, ptrIsSmaller)
	m.Unlock()
}

func (m *mapTree) Get(key interface{}) *avltree.Tree {
	return m.entries[createKey(key)]
}
