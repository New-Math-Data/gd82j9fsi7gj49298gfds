package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortestPathSameNode(t *testing.T) {
	bld := NewBuilder()
	a := bld.AddNode("A")
	b := bld.AddNode("B")
	bld.ConnectNodes(a, b)
	g := bld.Build()
	assert.Equal(t, 0, g.ShortestPath(a, a))
}

func TestShortestPathConnected(t *testing.T) {
	bld := NewBuilder()
	a, b, c := bld.AddNode("A"), bld.AddNode("B"), bld.AddNode("C")
	bld.ConnectNodes(a, b)
	bld.ConnectNodes(b, c)
	g := bld.Build()
	assert.Equal(t, 2, g.ShortestPath(a, c))
	assert.Equal(t, 1, g.ShortestPath(a, b))
}

func TestShortestPathDisconnected(t *testing.T) {
	bld := NewBuilder()
	a := bld.AddNode("A")
	bld.ConnectNodes(a, bld.AddNode("B"))
	c := bld.AddNode("C")
	bld.ConnectNodes(c, bld.AddNode("D"))
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
