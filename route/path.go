package route

type Path struct {
	Type       uint8
	StaticPath *StaticPath
	BGPPath    *BGPPath
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
