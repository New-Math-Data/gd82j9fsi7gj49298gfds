package graph

import (
	"fmt"
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

// Just to ensure truly handle deep graphs, not just a few hops.
func TestShortestPathConnectedDeep(t *testing.T) {
	bld := NewBuilder()
	source := bld.AddNode("source")

	// unrelated branch
	other := bld.AddNode("other")
	bld.ConnectNodes(source, other)

	// add 100 nodes between the source and target plus some diversion
	// subgraphs to make sure our BFS handles those.
	previous := source
	for i := 0; i < 100; i++ {
		intermediate := bld.AddNode(NodeID(fmt.Sprintf("IntermediateNode_%d", i)))
		bld.ConnectNodes(previous, intermediate)

		if i%2 == 0 {
			// Create a diversion subgraph to test that we don't get lost in our search:
			//
			//     other (initial inrelated branching)
			//    /
			// source -> ... -> intermediate -> intermediate -> target
			//                           \         \      \
			//              diversions:   \---------A      B
			//                                              \
			//                                               C
			diversionA := bld.AddNode(NodeID(fmt.Sprintf("DiversionNode_%da", i)))
			diversionB := bld.AddNode(NodeID(fmt.Sprintf("DiversionNode_%db", i)))
			diversionC := bld.AddNode(NodeID(fmt.Sprintf("DiversionNode_%dc", i)))
			bld.ConnectNodes(intermediate, diversionA)
			bld.ConnectNodes(intermediate, diversionB)
			bld.ConnectNodes(diversionB, diversionC)
			bld.ConnectNodes(previous, diversionA) // Note this connects to the source on the first iteration
		}

		previous = intermediate
	}

	target := bld.AddNode("target")
	bld.ConnectNodes(previous, target)

	g := bld.Build()

	// because there are 100 nodes *between* the source and target,
	// there's a final hop from the final intermediate node to the
	// target, so the total number of hops is 101, not 100.
	assert.Equal(t, 101, g.ShortestPath(source, target))
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
