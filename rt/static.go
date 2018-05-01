package rt

type StaticPath struct {
	NextHop uint32
}

func (r *Route) staticPathSelection() (best *Path, active []*Path) {
	if r.paths == nil {
		return nil, nil
	}

	if len(r.paths) == 0 {
		return nil, nil
	}

	for _, p := range r.paths {
		if p.Type != StaticPathType {
			continue
		}

		active = append(active, p)
		best = p
	}

	return
}
