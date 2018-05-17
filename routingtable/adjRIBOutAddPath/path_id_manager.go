package adjRIBOutAddPath

import (
	"fmt"
)

var maxUint32 = ^uint32(0)

// pathIDManager manages BGP path identifiers for add-path. This is no thread safe (and doesn't need to be).
type pathIDManager struct {
	ids  map[uint32]struct{}
	last uint32
	used uint32
}

func newPathIDManager() *pathIDManager {
	return &pathIDManager{
		ids: make(map[uint32]struct{}),
	}
}

func (fm *pathIDManager) getNewID() (uint32, error) {
	if fm.used == maxUint32 {
		return 0, fmt.Errorf("Out of path IDs")
	}

	fm.last++
	for {
		if _, exists := fm.ids[fm.last]; exists {
			fm.last++
			continue
		}
		break
	}

	ret := fm.last
	fm.used++

	return ret, nil
}

func (fm *pathIDManager) releaseID(id uint32) {
	if _, exists := fm.ids[id]; exists {
		delete(fm.ids, id)
		fm.used--
	}
}
