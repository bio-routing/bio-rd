package rt

type StaticPath struct {
	NextHop uint32
}

func (r *Route) staticPathSelection() (res []*Path) {
	if len(r.paths) == 1 {
		copy(res, r.paths)
		return res
	}

	for _, p := range r.paths {
		if p.Type != StaticPathType {
			continue
		}

		res = append(res, p)
	}

	return
}
