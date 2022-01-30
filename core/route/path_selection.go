package route

func (p *Path) compare(q *Path) int {
	if p.hasValidNexthop() && !q.hasValidNexthop() {
		return 1
	}

	if !p.hasValidNexthop() && q.hasValidNexthop() {
		return -1
	}

	if p.isFiltered() && !q.isFiltered() {
		return 1
	}

	if !p.isFiltered() && q.isFiltered() {
		return -1
	}

	return 0
}
