package route

import (
	"sync"

	"github.com/bio-routing/bio-rd/net"
)

// StaticPathType indicats a path is a static path
const StaticPathType = 1

// BGPPathType indicates a path is a BGP path
const BGPPathType = 2

// OSPFPathType indicates a path is an OSPF path
const OSPFPathType = 3

// ISISPathType indicates a path is an ISIS path
const ISISPathType = 4

// Route links a prefix to paths
type Route struct {
	pfx         net.Prefix
	mu          sync.Mutex
	bestPath    *Path
	activePaths []*Path
	paths       []*Path
}

// NewRoute generates a new route
func NewRoute(pfx net.Prefix, p *Path) *Route {
	r := &Route{
		pfx: pfx,
	}

	if p == nil {
		r.paths = make([]*Path, 0)
		return r
	}

	r.paths = []*Path{p}
	return r
}

// Prefix gets the prefix of route `r`
func (r *Route) Prefix() net.Prefix {
	return r.pfx
}

// Addr gets a routes address
func (r *Route) Addr() uint32 {
	return r.pfx.Addr()
}

// Pfxlen gets a routes prefix length
func (r *Route) Pfxlen() uint8 {
	return r.pfx.Pfxlen()
}

// Paths returns a copy of the list of paths associated with route r
func (r *Route) Paths() []*Path {
	if r.paths == nil {
		return nil
	}

	ret := make([]*Path, len(r.paths))
	copy(ret, r.paths)
	return ret
}

// AddPath adds path p to route r
func (r *Route) AddPath(p *Path) {
	if p == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.paths = append(r.paths, p)
}

// RemovePath removes path `p` from route `r`. Returns length of path list after removing path `p`
func (r *Route) RemovePath(p *Path) int {
	if p == nil {
		return len(r.paths)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.paths = removePath(r.paths, p)
	return len(r.paths)
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
