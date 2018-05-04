package route

// BGPPath represents a set of BGP path attributes
type BGPPath struct {
	PathIdentifier uint32
	NextHop        uint32
	LocalPref      uint32
	ASPath         string
	ASPathLen      uint16
	Origin         uint8
	MED            uint32
	EBGP           bool
	BGPIdentifier  uint32
	Source         uint32
}

func (r *Route) bgpPathSelection() (best *Path, active []*Path) {
	// TODO: Implement next hop lookup and compare IGP metrics
	for _, p := range r.paths {
		if p.Type != BGPPathType {
			continue
		}

		if len(active) == 0 {
			active = append(active, p)
			best = p
			continue
		}

		if active[0].BGPPath.ecmp(p.BGPPath) {
			active = append(active, p)
			if !r.bestPath.BGPPath.better(p.BGPPath) {
				continue
			}

			best = p
			continue
		}

		if !active[0].BGPPath.betterECMP(p.BGPPath) {
			continue
		}

		active = []*Path{p}
		best = p
	}

	return best, active
}

func (b *BGPPath) betterECMP(c *BGPPath) bool {
	if c.LocalPref < b.LocalPref {
		return false
	}

	if c.LocalPref > b.LocalPref {
		return true
	}

	if c.ASPathLen > b.ASPathLen {
		return false
	}

	if c.ASPathLen < b.ASPathLen {
		return true
	}

	if c.Origin > b.Origin {
		return false
	}

	if c.Origin < b.Origin {
		return true
	}

	if c.MED > b.MED {
		return false
	}

	if c.MED < b.MED {
		return true
	}

	return false
}

func (b *BGPPath) better(c *BGPPath) bool {
	if b.betterECMP(c) {
		return true
	}

	if c.BGPIdentifier < b.BGPIdentifier {
		return true
	}

	if c.Source < b.Source {
		return true
	}

	return false
}

func (b *BGPPath) ecmp(c *BGPPath) bool {
	return b.LocalPref == c.LocalPref && b.ASPathLen == c.ASPathLen && b.Origin == c.Origin && b.MED == c.MED
}
