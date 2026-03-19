package graph

// Graph builder allows you to "stream" nodes and edges into the data structure
// instead of collect and manage all that yourself.
type Builder struct {
	nodes map[NodeID]*Node
	edges []*Edge
}

func NewBuilder() *Builder {
	return &Builder{nodes: make(map[NodeID]*Node)}
}

func (b *Builder) AddNode(id NodeID) *Node {
	if n, ok := b.nodes[id]; ok {
		return n
	}
	n := &Node{ID: id}
	b.nodes[id] = n
	return n
}

func (b *Builder) AddEdge(from, to NodeID) *Edge {
	f := b.AddNode(from)
	t := b.AddNode(to)
	e := &Edge{From: f, To: t}
	b.edges = append(b.edges, e)
	return e
}

func (b *Builder) Build() *Graph {
	adj := make(map[NodeID][]*Node, len(b.nodes))
	for _, e := range b.edges {
		adj[e.From.ID] = append(adj[e.From.ID], e.To)
	}
	return &Graph{nodes: b.nodes, edges: b.edges, adj: adj}
}
