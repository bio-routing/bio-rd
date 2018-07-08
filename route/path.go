package route

import (
	"fmt"
)

type Path struct {
	Type       uint8
	StaticPath *StaticPath
	BGPPath    *BGPPath
}

// Compare returns negative if p < q, 0 if paths are equal, positive if p > q
func (p *Path) Compare(q *Path) int8 {
	switch {
	case p == nil && q == nil:
		return 0
	case p == nil:
		return -1
	case q == nil:
		return 1
	default:
	}

	if p.Type > q.Type {
		return 1
	}

	if p.Type < q.Type {
		return -1
	}

	switch p.Type {
	case BGPPathType:
		return p.BGPPath.Compare(q.BGPPath)
	case StaticPathType:
		return p.StaticPath.Compare(q.StaticPath)
	}

	panic("Unknown path type")
}

func (p *Path) ECMP(q *Path) bool {
	switch p.Type {
	case BGPPathType:
		return p.BGPPath.ECMP(q.BGPPath)
	case StaticPathType:
		return p.StaticPath.ECMP(q.StaticPath)
	}

	panic("Unknown path type")
}

func (p *Path) Equal(q *Path) bool {
	if p == nil || q == nil {
		return false
	}
	return p.Compare(q) == 0
}

// PathsDiff gets the list of elements contained by a but not b
func PathsDiff(a, b []*Path) []*Path {
	ret := make([]*Path, 0)

	for _, pa := range a {
		if !pathsContains(pa, b) {
			ret = append(ret, pa)
		}
	}

	return ret
}

func pathsContains(needle *Path, haystack []*Path) bool {
	for _, p := range haystack {
		if p == needle {
			return true
		}
	}

	return false
}

func (p *Path) Print() string {
	protocol := ""
	switch p.Type {
	case StaticPathType:
		protocol = "static"
	case BGPPathType:
		protocol = "BGP"
	}

	ret := fmt.Sprintf("\tProtocol: %s\n", protocol)
	switch p.Type {
	case StaticPathType:
		ret += "Not implemented yet"
	case BGPPathType:
		ret += p.BGPPath.Print()
	}

	return ret
}

func (p *Path) Copy() *Path {
	if p == nil {
		return nil
	}

	cp := *p
	cp.BGPPath = cp.BGPPath.Copy()
	cp.StaticPath = cp.StaticPath.Copy()

	return &cp
}
