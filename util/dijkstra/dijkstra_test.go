package dijkstra

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSPT(t *testing.T) {
	tests := []struct {
		name     string
		nodes    []Node
		edges    []Edge
		expected *SPT
	}{
		{
			name: "Test #1",
			nodes: []Node{
				{
					Name: "A",
				},
				{
					Name: "B",
				}, {
					Name: "C",
				}, {
					Name: "D",
				},
			},
			edges: []Edge{
				{
					NodeA:    Node{Name: "A"},
					NodeB:    Node{Name: "B"},
					Distance: 1,
				},
				{
					NodeA:    Node{Name: "B"},
					NodeB:    Node{Name: "A"},
					Distance: 1,
				},
				{
					NodeA:    Node{Name: "B"},
					NodeB:    Node{Name: "C"},
					Distance: 5,
				},
				{
					NodeA:    Node{Name: "C"},
					NodeB:    Node{Name: "B"},
					Distance: 5,
				},
				{
					NodeA:    Node{Name: "B"},
					NodeB:    Node{Name: "D"},
					Distance: 1,
				},
				{
					NodeA:    Node{Name: "D"},
					NodeB:    Node{Name: "B"},
					Distance: 1,
				},
				{
					NodeA:    Node{Name: "A"},
					NodeB:    Node{Name: "D"},
					Distance: 5,
				},
				{
					NodeA:    Node{Name: "D"},
					NodeB:    Node{Name: "A"},
					Distance: 5,
				},
				{
					NodeA:    Node{Name: "D"},
					NodeB:    Node{Name: "C"},
					Distance: 5,
				},
				{
					NodeA:    Node{Name: "C"},
					NodeB:    Node{Name: "D"},
					Distance: 5,
				},
			},
			expected: &SPT{
				Node{Name: "A"}: Path{
					Edges:    []Edge{},
					Distance: 0,
				},
				Node{Name: "B"}: Path{
					Edges: []Edge{
						{
							NodeA:    Node{Name: "A"},
							NodeB:    Node{Name: "B"},
							Distance: 1,
						},
					},
					Distance: 1,
				},
				Node{Name: "C"}: Path{
					Edges: []Edge{
						{
							NodeA:    Node{Name: "A"},
							NodeB:    Node{Name: "B"},
							Distance: 1,
						},
						{
							NodeA:    Node{Name: "B"},
							NodeB:    Node{Name: "D"},
							Distance: 1,
						},
						{
							NodeA:    Node{Name: "D"},
							NodeB:    Node{Name: "C"},
							Distance: 1,
						},
					},
					Distance: 3,
				},
				Node{Name: "D"}: Path{
					Edges: []Edge{
						{
							NodeA:    Node{Name: "A"},
							NodeB:    Node{Name: "B"},
							Distance: 1,
						},
						{
							NodeA:    Node{Name: "B"},
							NodeB:    Node{Name: "D"},
							Distance: 1,
						},
					},
					Distance: 2,
				},
			},
		},
	}

	for _, test := range tests {
		top := NewTopologay(test.nodes, test.edges)
		spt := top.SPT(Node{Name: "A"})

		assert.Equalf(t, test.expected, spt, "Test %q", test.name)
	}
}
