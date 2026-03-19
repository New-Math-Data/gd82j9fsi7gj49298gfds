package graph

type NodeID string

type Node struct {
	ID NodeID
}

type Edge struct {
	From *Node
	To   *Node
}

type Graph struct {
	nodes map[NodeID]*Node
	edges []*Edge
	adj   map[NodeID][]*Node
}

func NewGraph(nodes []*Node, edges []*Edge) *Graph {
	nodeMap := make(map[NodeID]*Node, len(nodes))
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}
	adj := make(map[NodeID][]*Node, len(nodes))
	for _, e := range edges {
		adj[e.From.ID] = append(adj[e.From.ID], e.To)
	}
	return &Graph{nodes: nodeMap, edges: edges, adj: adj}
}

func (g *Graph) Nodes() []*Node {
	nodes := make([]*Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

func (g *Graph) Edges() []*Edge {
	return g.edges
}

func (g *Graph) Node(id NodeID) *Node {
	return g.nodes[id]
}

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
