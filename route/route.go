package route

import (
	"fmt"
	"sort"
	"sync"

	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route/api"
)

const (
	_ = iota // 0

	// StaticPathType indicators a path is a static path
	StaticPathType

	// BGPPathType indicates a path is a BGP path
	BGPPathType

	// OSPFPathType indicates a path is an OSPF path
	OSPFPathType

	// ISISPathType indicates a path is an ISIS path
	ISISPathType

	// FIBPathType indicates a path is a FIB path
	FIBPathType
)

// Route links a prefix to paths
type Route struct {
	pfx       *net.Prefix
	mu        sync.Mutex
	paths     []*Path
	ecmpPaths uint
}

// NewRoute generates a new route with path p
func NewRoute(pfx *net.Prefix, p *Path) *Route {
	r := &Route{
		pfx: pfx,
	}

	if p == nil {
		r.paths = make([]*Path, 0, 2)
		return r
	}

	r.paths = append(r.paths, p)
	return r
}

// NewRouteAddPath generates a new route with paths p
func NewRouteAddPath(pfx *net.Prefix, p []*Path) *Route {
	r := &Route{
		pfx: pfx,
	}

	if p == nil {
		r.paths = make([]*Path, 0)
		return r
	}

	r.paths = append(r.paths, p...)

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
func (r *Route) Prefix() *net.Prefix {
	return r.pfx
}

// Addr gets a routes address
func (r *Route) Addr() *net.IP {
	return r.pfx.Addr().Ptr()
}

// Pfxlen gets a routes prefix length
func (r *Route) Pfxlen() uint8 {
	return r.pfx.Len()
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

// ReplacePath replace path old with new
func (r *Route) ReplacePath(old *Path, new *Path) error {
	for i := range r.paths {
		if r.paths[i].Equal(old) {
			r.paths[i] = new
			return nil
		}
	}

	return fmt.Errorf("Path not found")
}

func removePath(paths []*Path, remove *Path) []*Path {
	i := -1
	for j := range paths {
		if paths[j].Compare(remove) {
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
		return r.paths[i].Select(r.paths[j]) == 1
	})

	r.updateEqualPathCount()
}

// Equal compares if two routes are the same
func (r *Route) Equal(other *Route) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	pfxEqual := r.pfx.Equal(other.pfx)
	ecmpPathsEqual := r.ecmpPaths == other.ecmpPaths
	pathsEqual := comparePathSlice(r.paths, other.paths)

	return pfxEqual && ecmpPathsEqual && pathsEqual
}

// Compare two path pointer slices if they are equal
func comparePathSlice(left, right []*Path) bool {
	if left == nil && right == nil {
		return true
	}

	if len(left) != len(right) {
		return false
	}

	for _, leftPath := range left {
		if !compareItemExists(leftPath, right) {
			return false
		}
	}

	return true
}

func compareItemExists(needle *Path, haystack []*Path) bool {
	for _, compare := range haystack {
		if needle.Equal(compare) {
			return true
		}
	}

	return false
}

func (r *Route) GetBGPOriginatingAS() *uint32 {
	lastASPathSeg := r.BestPath().BGPPath.ASPath.GetLastSequenceSegment()
	if lastASPathSeg != nil {
		origASN := lastASPathSeg.GetLastASN()
		if origASN != nil {
			return origASN
		}
	}
	return nil
}

func (r *Route) IsBGPOriginatedBy(asn uint32) bool {
	if orig := r.GetBGPOriginatingAS(); orig != nil {
		return *r.GetBGPOriginatingAS() == asn
	}
	return false
}

// ToProto converts route to proto route
func (r *Route) ToProto() *api.Route {
	a := &api.Route{
		Pfx:   r.pfx.ToProto(),
		Paths: make([]*api.Path, len(r.paths)),
	}

	for i := range r.paths {
		a.Paths[i] = r.paths[i].ToProto()
	}

	return a
}

// RouteFromProtoRoute converts a proto Route to a Route
func RouteFromProtoRoute(ar *api.Route, dedup bool) *Route {
	r := &Route{
		pfx:   net.NewPrefixFromProtoPrefix(ar.Pfx),
		paths: make([]*Path, 0, len(ar.Paths)),
	}

	for i := range ar.Paths {
		p := &Path{}
		switch ar.Paths[i].Type {
		case api.Path_BGP:
			p.Type = BGPPathType
			p.BGPPath = BGPPathFromProtoBGPPath(ar.Paths[i].BgpPath, dedup)
		case api.Path_Static:
			p.Type = StaticPathType
			p.StaticPath = StaticPathFromProtoStaticPath(ar.Paths[i].StaticPath, dedup)
		}

		r.paths = append(r.paths, p)
	}

	return r
}

func (r *Route) updateEqualPathCount() {
	if len(r.paths) == 0 {
		r.ecmpPaths = 0
		return
	}

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

// Print returns a printable representation of route `r`
func (r *Route) Print() string {
	ret := fmt.Sprintf("%s:\n", r.pfx.String())
	ret += "All Paths:\n"
	for _, p := range r.paths {
		ret += p.Print()
	}

	return ret
}
