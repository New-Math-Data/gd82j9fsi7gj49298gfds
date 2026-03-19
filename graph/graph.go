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
type Graph struct {
	nodes map[NodeID]*Node
	edges []*Edge
	adj   map[NodeID][]*Node // Adjacency map allows O(1) lookups of a given node to its connected nodes
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

	type item struct {
		node *Node
		dist int
	}
	visited := map[NodeID]bool{source.ID: true}

	queue := []item{{source, 0}}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, neighbor := range g.adj[curr.node.ID] {
			if neighbor == target {
				return curr.dist + 1
			}
			if !visited[neighbor.ID] {
				visited[neighbor.ID] = true
				queue = append(queue, item{neighbor, curr.dist + 1})
			}
		}
	}

	return -1
}
