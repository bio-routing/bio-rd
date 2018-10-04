package dijkstra

type Topology struct {
	nodes []Node
	edges map[Node]map[Node]int64
}

type Node struct {
	Name string
}

type Edge struct {
	NodeA    Node
	NodeB    Node
	Distance int64
}

type SPT map[Node]Path

type Path struct {
	Edges    []Edge
	Distance int64
}

// NewTopologay creates a new topology
func NewTopologay(nodes []Node, edges []Edge) *Topology {
	t := &Topology{
		nodes: nodes,
		edges: make(map[Node]map[Node]int64),
	}

	for _, e := range edges {
		if _, ok := t.edges[e.NodeA]; !ok {
			t.edges[e.NodeA] = make(map[Node]int64)
		}

		t.edges[e.NodeA][e.NodeB] = e.Distance
	}

	return t
}

// SPT calculates the shortest path tree
func (t *Topology) SPT(from Node) *SPT {
	spt := make(SPT)

	for _, n := range t.nodes {
		spt[n] = Path{
			Edges:    make([]Edge, 0),
			Distance: -1,
		}
	}

	f := spt[from]
	f.Distance = 0
	spt[from] = f

	baseDistance := int64(0)
	for n, d := range t.edges[from] {
		nPath := spt[n]
		if nPath.Distance < 0 {
			nPath.Distance = d
			nPath.Edges = []Edge{
				{
					NodeA:    from,
					NodeB:    n,
					Distance: d,
				},
			}

			spt[n] = nPath
			continue
		}

		if spt[n].Distance < baseDistance+d {
			continue
		}

		nPath.Distance += d
		nPath.Edges = []Edge{
			{
				NodeA:    from,
				NodeB:    n,
				Distance: d,
			},
		}

	}

	return &spt
}
