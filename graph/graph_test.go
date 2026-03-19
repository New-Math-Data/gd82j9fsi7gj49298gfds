package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortestPathSameNode(t *testing.T) {
	bld := NewBuilder()
	a := bld.AddNode("A")
	bld.AddEdge("A", "B")
	bld.AddEdge("B", "A")
	g := bld.Build()
	assert.Equal(t, 0, g.ShortestPath(a, a))
}

func TestShortestPathConnected(t *testing.T) {
	bld := NewBuilder()
	a, b, c := bld.AddNode("A"), bld.AddNode("B"), bld.AddNode("C")
	bld.AddEdge("A", "B")
	bld.AddEdge("B", "A")
	bld.AddEdge("B", "C")
	bld.AddEdge("C", "B")
	g := bld.Build()
	assert.Equal(t, 2, g.ShortestPath(a, c))
	assert.Equal(t, 1, g.ShortestPath(a, b))
}

func TestShortestPathDisconnected(t *testing.T) {
	bld := NewBuilder()
	a, c := bld.AddNode("A"), bld.AddNode("C")
	bld.AddEdge("A", "B")
	bld.AddEdge("B", "A")
	bld.AddEdge("C", "D")
	bld.AddEdge("D", "C")
	g := bld.Build()
	assert.Equal(t, -1, g.ShortestPath(a, c))
}

func TestShortestPathNilNode(t *testing.T) {
	bld := NewBuilder()
	a := bld.AddNode("A")
	g := bld.Build()
	assert.Equal(t, -1, g.ShortestPath(a, nil))
	assert.Equal(t, -1, g.ShortestPath(nil, a))
}
