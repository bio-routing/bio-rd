package route

// StaticPath represents a static path of a route
type StaticPath struct {
	NextHop uint32
}

func (r *Route) staticPathSelection() {
	if len(r.paths) == 0 {
		return
	}

	r.ecmpPaths = uint(len(r.paths))
	return
}

// Compare returns negative if s < t, 0 if paths are equal, positive if s > t
func (s *StaticPath) Compare(t *StaticPath) int8 {
	return 0
}

// ECMP determines if path s and t are equal in terms of ECMP
func (s *StaticPath) ECMP(t *StaticPath) bool {
	return true
}
