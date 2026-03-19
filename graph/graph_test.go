package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortestPathSameNode(t *testing.T) {
	a, b := &Node{ID: "A"}, &Node{ID: "B"}
	g := NewGraph([]*Node{a, b}, []*Edge{{a, b}, {b, a}})
	assert.Equal(t, 0, g.ShortestPath(a, a))
}

func TestShortestPathConnected(t *testing.T) {
	a, b, c := &Node{ID: "A"}, &Node{ID: "B"}, &Node{ID: "C"}
	g := NewGraph([]*Node{a, b, c}, []*Edge{{a, b}, {b, a}, {b, c}, {c, b}})
	assert.Equal(t, 2, g.ShortestPath(a, c))
	assert.Equal(t, 1, g.ShortestPath(a, b))
}

func TestShortestPathDisconnected(t *testing.T) {
	a, b, c, d := &Node{ID: "A"}, &Node{ID: "B"}, &Node{ID: "C"}, &Node{ID: "D"}
	g := NewGraph([]*Node{a, b, c, d}, []*Edge{{a, b}, {b, a}, {c, d}, {d, c}})
	assert.Equal(t, -1, g.ShortestPath(a, c))
}

func TestShortestPathNilNode(t *testing.T) {
	a := &Node{ID: "A"}
	g := NewGraph([]*Node{a}, nil)
	assert.Equal(t, -1, g.ShortestPath(a, nil))
	assert.Equal(t, -1, g.ShortestPath(nil, a))
}
