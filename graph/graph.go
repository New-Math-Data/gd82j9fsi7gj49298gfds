package graph

type NodeID string

type Node struct {
	ID NodeID
}

type Edge struct {
	A *Node
	B *Node
}

// Represents an undirected and unweighted graph.
// Create with graph.Builder.
type Graph struct {
	nodes map[NodeID]*Node
	edges []*Edge

	// Adjacency map allows O(1) lookups of a given node to its connected nodes
	// at the memory cost of node pointer duplication.
	adj map[NodeID][]*Node
}

func (g *Graph) Edges() []*Edge {
	return g.edges
}

func (g *Graph) Node(id NodeID) *Node {
	return g.nodes[id]
}

// Breadth-first-searh for the smallest number of hops between two nodes.
// Returns -1 if the two nodes are unconnected.
func (g *Graph) ShortestPath(source, target *Node) int {
	if source == nil || target == nil {
		return -1
	}

	if source == target {
		return 0
	}

	// Keep up with which nodes have already been visited and
	// how far away they are from the source node.
	// Store visited nodes in a map for O(1) lookups.
	type item struct {
		node *Node
		dist int
	}
	visited := map[NodeID]bool{source.ID: true}

	// "Flatten" the recursive search by queueing up the parameters of
	// each recursive call. Prevents stack overflow by storing stuff
	// in the heap (which is much larger than the stack).
	queue := []item{{source, 0}}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		// Look at all the neighbors of the current node.
		for _, neighbor := range g.adj[curr.node.ID] {
			if neighbor == target {
				// We found the target!
				return curr.dist + 1
			}

			// Okay, so we haven't found the target yet, so we need to
			// search further in the graph.
			// If we've already visited this node, skip over. Prevents search cycles.
			// But if we haven't, visit the this node and search its
			// neighbors (so, increase the distance).
			// This would be the recursive search if we weren't
			// queueing up the parameters to flatten the recursion.
			if !visited[neighbor.ID] {
				visited[neighbor.ID] = true
				queue = append(queue, item{neighbor, curr.dist + 1})
			}
		}
	}

	return -1
}
