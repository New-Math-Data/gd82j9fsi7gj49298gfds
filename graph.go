package main

type Graph struct {
	edges map[string][]string
}

func BuildGraph(collection *GeoJSONFeatureCollection) *Graph {
	g := &Graph{edges: make(map[string][]string)}
	for _, f := range collection.Features {
		if f.Properties.AssetType != "Line" {
			continue
		}
		ca := f.Properties.ConnectedAssets
		if len(ca.Sources) == 0 || len(ca.Targets) == 0 {
			continue
		}
		bus1 := ca.Sources[0]
		bus2 := ca.Targets[0]
		g.edges[bus1] = append(g.edges[bus1], bus2)
		g.edges[bus2] = append(g.edges[bus2], bus1)
	}
	return g
}

func (g *Graph) ShortestPath(source, target string) int {
	if source == target {
		return 0
	}
	type item struct {
		node string
		dist int
	}
	visited := map[string]bool{source: true}
	queue := []item{{source, 0}}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, neighbor := range g.edges[curr.node] {
			if neighbor == target {
				return curr.dist + 1
			}
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, item{neighbor, curr.dist + 1})
			}
		}
	}
	return -1
}
