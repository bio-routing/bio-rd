package adjRIBOut

import (
	"fmt"

	"github.com/bio-routing/bio-rd/route"
)

var maxUint32 = ^uint32(0)

// pathIDManager manages BGP path identifiers for add-path. This is no thread safe (and doesn't need to be).
type pathIDManager struct {
	ids      map[uint32]uint64
	idByPath map[string]uint32
	last     uint32
	used     uint32
}

func newPathIDManager() *pathIDManager {
	return &pathIDManager{
		ids:      make(map[uint32]uint64),
		idByPath: make(map[string]uint32),
	}
}

func (fm *pathIDManager) addPath(p *route.Path) (uint32, error) {
	hash := p.BGPPath.ComputeHash()

	if _, exists := fm.idByPath[hash]; exists {
		id := fm.idByPath[hash]
		fm.ids[id]++
		return id, nil
	}

	if fm.used == maxUint32 {
		return 0, fmt.Errorf("out of path IDs")
	}

	fm.last++
	for {
		if _, exists := fm.ids[fm.last]; exists {
			fm.last++
			continue
		}
		break
	}

	fm.idByPath[hash] = fm.last
	fm.ids[fm.last] = 1
	fm.used++

	return fm.last, nil
}

func (fm *pathIDManager) releasePath(p *route.Path) (uint32, error) {
	hash := p.BGPPath.ComputeHash()

	if _, exists := fm.idByPath[hash]; !exists {
		return 0, fmt.Errorf("ID not found for path: %s", p.Print())
	}

	id := fm.idByPath[hash]
	fm.ids[id]--
	if fm.ids[id] == 0 {
		delete(fm.ids, fm.idByPath[hash])
		delete(fm.idByPath, hash)
	}
	fm.used--

	return id, nil
}
