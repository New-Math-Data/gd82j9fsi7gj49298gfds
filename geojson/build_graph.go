package geojson

import "opendss-assessment/graph"

func (c *GeoJSONFeatureCollection) BuildGraph() *graph.Graph {
	nodeMap := make(map[graph.NodeID]*graph.Node)
	getOrCreate := func(id graph.NodeID) *graph.Node {
		if n, ok := nodeMap[id]; ok {
			return n
		}
		n := &graph.Node{ID: graph.NodeID(id)}
		nodeMap[id] = n
		return n
	}

	var edges []*graph.Edge
	for _, f := range c.Features {
		if f.Properties.AssetType != "Line" {
			continue
		}
		ca := f.Properties.ConnectedAssets
		if len(ca.Sources) == 0 || len(ca.Targets) == 0 {
			continue
		}
		n1 := getOrCreate(graph.NodeID(ca.Sources[0]))
		n2 := getOrCreate(graph.NodeID(ca.Targets[0]))
		edges = append(edges, &graph.Edge{From: n1, To: n2})
		edges = append(edges, &graph.Edge{From: n2, To: n1})
	}

	nodes := make([]*graph.Node, 0, len(nodeMap))
	for _, n := range nodeMap {
		nodes = append(nodes, n)
	}
	return graph.NewGraph(nodes, edges)
}
