package route

import (
	"fmt"
	"sort"
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
	pfx       net.Prefix
	mu        sync.Mutex
	paths     []*Path
	ecmpPaths uint
}

// NewRoute generates a new route with path p
func NewRoute(pfx net.Prefix, p *Path) *Route {
	r := &Route{
		pfx: pfx,
	}

	if p == nil {
		r.paths = make([]*Path, 0)
		return r
	}

	r.paths = append(r.paths, p)
	return r
}

// NewRouteAddPath generates a new route with paths p
func NewRouteAddPath(pfx net.Prefix, p []*Path) *Route {
	r := &Route{
		pfx: pfx,
	}

	if p == nil {
		r.paths = make([]*Path, 0)
		return r
	}

	for _, path := range p {
		r.paths = append(r.paths, path)
	}
	return r
}

// Copy returns a copy of route r
func (r *Route) Copy() *Route {
	if r == nil {
		return nil
	}
	n := &Route{
		pfx:       r.pfx,
		ecmpPaths: r.ecmpPaths,
	}
	n.paths = make([]*Path, len(r.paths))
	copy(n.paths, r.paths)
	return n
}

// Prefix gets the prefix of route `r`
func (r *Route) Prefix() net.Prefix {
	return r.pfx
}

// Addr gets a routes address
func (r *Route) Addr() net.IP {
	return r.pfx.Addr()
}

// Pfxlen gets a routes prefix length
func (r *Route) Pfxlen() uint8 {
	return r.pfx.Pfxlen()
}

// Paths returns a copy of the list of paths associated with route r
func (r *Route) Paths() []*Path {
	if r == nil || r.paths == nil {
		return nil
	}

	ret := make([]*Path, len(r.paths))
	copy(ret, r.paths)
	return ret
}

// ECMPPathCount returns the count of ecmp paths for route r
func (r *Route) ECMPPathCount() uint {
	if r == nil {
		return 0
	}
	return r.ecmpPaths
}

// ECMPPaths returns a copy of the list of paths associated with route r
func (r *Route) ECMPPaths() []*Path {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.paths) == 0 {
		return nil
	}

	ret := make([]*Path, r.ecmpPaths)
	copy(ret, r.paths[0:r.ecmpPaths])
	return ret
}

// BestPath returns the current best path. nil if non exists
func (r *Route) BestPath() *Path {
	if r == nil {
		return nil
	}
	if len(r.paths) == 0 {
		return nil
	}

	return r.paths[0]
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

// PathSelection recalculates the best path + active paths
func (r *Route) PathSelection() {
	r.mu.Lock()
	defer r.mu.Unlock()

	sort.Slice(r.paths, func(i, j int) bool {
		return r.paths[i].Compare(r.paths[j]) == -1
	})

	r.updateEqualPathCount()
}

func (r *Route) updateEqualPathCount() {
	count := uint(1)
	for i := 0; i < len(r.paths)-1; i++ {
		if !r.paths[i].ECMP(r.paths[i+1]) {
			break
		}
		count++
	}

	r.ecmpPaths = count
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

// Print returns a prinatble representation of route `r`
func (r *Route) Print() string {
	ret := fmt.Sprintf("%s:\n", r.pfx.String())
	ret += fmt.Sprintf("All Paths:\n")
	for _, p := range r.paths {
		ret += p.Print()
	}

	return ret
}
