package dijkstra

// Topology represents a network topology
type Topology struct {
	nodes map[Node]int64
	edges map[Node]map[Node]int64
}

// Node represents a node in a graph
type Node struct {
	Name string
}

// Edge represents a directed edge in a graph
type Edge struct {
	NodeA    Node
	NodeB    Node
	Distance int64
}

// SPT represents a shortest path tree
type SPT map[Node]Path

// Path represents a path through a graph
type Path struct {
	Edges    []Edge
	Distance int64
}

// NewTopology creates a new topology
func NewTopologay(nodes []Node, edges []Edge) *Topology {
	t := &Topology{
		nodes: make(map[Node]int64),
		edges: make(map[Node]map[Node]int64),
	}

	for _, n := range nodes {
		t.nodes[n] = -1
	}

	for _, e := range edges {
		if _, ok := t.edges[e.NodeA]; !ok {
			t.edges[e.NodeA] = make(map[Node]int64)
		}

		t.edges[e.NodeA][e.NodeB] = e.Distance
	}

	return t
}

func (t *Topology) newSPT() SPT {
	spt := make(SPT)

	for n := range t.nodes {
		spt[n] = Path{
			Edges:    make([]Edge, 0),
			Distance: -1,
		}
	}

	return spt
}

// SPT calculates the shortest path tree
func (t *Topology) SPT(from Node) SPT {
	spt := t.newSPT()

	tmp := spt[from]
	tmp.Distance = 0
	spt[from] = tmp

	unmarked := make(map[Node]struct{})
	for n := range t.nodes {
		if n == from {
			continue
		}
		unmarked[n] = struct{}{}
	}

	for len(unmarked) > 0 {
		for neighbor, distance := range t.edges[from] {
			if spt[neighbor].Distance == -1 {
				tmp := spt[neighbor]
				tmp.Distance = spt[from].Distance + distance
				tmp.Edges = make([]Edge, len(spt[from].Edges)+1)
				copy(tmp.Edges, spt[from].Edges)
				tmp.Edges[len(spt[from].Edges)] = Edge{
					NodeA:    from,
					NodeB:    neighbor,
					Distance: distance,
				}
				spt[neighbor] = tmp
				continue
			}

			if spt[from].Distance+distance < spt[neighbor].Distance {
				tmp := spt[neighbor]
				tmp.Distance = spt[from].Distance + distance
				tmp.Edges = make([]Edge, len(spt[from].Edges)+1)
				copy(tmp.Edges, spt[from].Edges)
				tmp.Edges[len(spt[from].Edges)] = Edge{
					NodeA:    from,
					NodeB:    neighbor,
					Distance: distance,
				}
				spt[neighbor] = tmp
				continue
			}
		}

		var next *Node
		nextDistance := int64(0)
		for candidate := range unmarked {
			if spt[candidate].Distance == -1 {
				continue
			}

			if next == nil {
				tmp := candidate
				next = &tmp
				nextDistance = spt[candidate].Distance
				continue
			}

			if spt[candidate].Distance < nextDistance {
				tmp := candidate
				next = &tmp
				nextDistance = spt[candidate].Distance
				continue
			}
		}

		from = *next
		delete(unmarked, from)
	}

	return spt
}
