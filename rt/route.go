package rt

import (
	net "github.com/bio-routing/bio-rd/net"
)

// Path Types
const StaticPathType = 1
const BGPPathType = 2
const OSPFPathType = 3
const ISISPathType = 4

type Path struct {
	Type       uint8
	StaticPath *StaticPath
	BGPPath    *BGPPath
}

type Route struct {
	pfx         *net.Prefix
	bestPath    *Path
	activePaths []*Path
	paths       []*Path
}

func NewRoute(pfx *net.Prefix, paths []*Path) *Route {
	r := &Route{
		pfx:   pfx,
		paths: paths,
	}

	return r
}

func (r *Route) Pfxlen() uint8 {
	return r.pfx.Pfxlen()
}

func (r *Route) Prefix() *net.Prefix {
	return r.pfx
}

func (r *Route) Remove(rm *Route) (final bool) {
	for _, del := range rm.paths {
		r.paths = removePath(r.paths, del)
	}

	return len(r.paths) == 0
}

// returns a list of Paths that are in a but not in b
func missingPaths(a, b []*Path) []*Path {
	ret := make([]*Path, 0)
	for _, p := range a {
		found := false
		for _, q := range b {
			if *p == *q {
				found = true
				break
			}
		}
		if !found {
			ret = append(ret, p)
		}
	}

	return ret
}

func removePath(paths []*Path, remove *Path) []*Path {
	i := -1
	for j := range paths {
		if paths[j].Equal(remove) {
			i = j
			break
		}
	}

	if i < 0 {
		return paths
	}

	copy(paths[i:], paths[i+1:])
	return paths[:len(paths)-1]
}

func (r *Route) removeAllPaths() {
	r.paths = make([]*Path, 0)
}

func (p *Path) Equal(q *Path) bool {
	if p == nil || q == nil {
		return false
	}

	if p.Type != q.Type {
		return false
	}

	switch p.Type {
	case BGPPathType:
		if *p.BGPPath != *q.BGPPath {
			return false
		}
	}

	return true
}

func (r *Route) AddPath(p *Path) {
	r.paths = append(r.paths, p)
	r.bestPaths()
}

func (r *Route) AddPaths(paths []*Path) {
	for _, p := range paths {
		r.paths = append(r.paths, p)
	}
	r.bestPaths()
}

func (r *Route) bestPaths() {
	var best *Path
	var active []*Path
	protocol := getBestProtocol(r.paths)

	switch protocol {
	case StaticPathType:
		best, active = r.staticPathSelection()
	case BGPPathType:
		best, active = r.bgpPathSelection()
	}

	r.bestPath = best
	r.activePaths = active
}

func getBestProtocol(paths []*Path) uint8 {
	best := uint8(0)
	for _, p := range paths {
		if best == 0 {
			best = p.Type
			continue
		}

		if p.Type < best {
			best = p.Type
		}
	}

	return best
}
