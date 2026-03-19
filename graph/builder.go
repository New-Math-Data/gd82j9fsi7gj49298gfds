package graph

// Graph builder allows you to "stream" nodes and edges into the data structure
// instead of collect and manage all that yourself.
type Builder struct {
	g *Graph
}

func NewBuilder() *Builder {
	return &Builder{g: &Graph{
		nodes: make(map[NodeID]*Node),
		adj:   make(map[NodeID][]*Node),
	}}
}

func (b *Builder) AddNode(id NodeID) *Node {
	if n, ok := b.g.nodes[id]; ok {
		return n
	}
	n := &Node{ID: id}
	b.g.nodes[id] = n
	return n
}

func (b *Builder) ConnectNodes(an, bn *Node) *Edge {
	e := &Edge{A: an, B: bn}
	b.g.edges = append(b.g.edges, e)
	b.g.adj[an.ID] = append(b.g.adj[an.ID], bn)
	b.g.adj[bn.ID] = append(b.g.adj[bn.ID], an)
	return e
}

func (b *Builder) Build() *Graph {
	return b.g
}
