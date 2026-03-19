package geojson

import "opendss-assessment/graph"

// Converts the GeoJSONFeatureCollection into an abstract Graph we can perform operations on.
func (c *FeatureCollection) BuildGraph() *graph.Graph {
	b := graph.NewBuilder()
	for _, f := range c.Features {
		// We only need to look at Lines to build the graph because when the FeatureCollection
		// was built, we already computed the connectivity between Lines and Buses.
		if f.Properties.AssetType != "Line" {
			continue
		}

		// A valid line has a single source and target....
		ca := f.Properties.ConnectedAssets
		if len(ca.Sources) == 0 || len(ca.Targets) == 0 {
			continue
		}

		// ...and we want the connection between them to be undirected and unweighted.
		src := b.AddNode(graph.NodeID(ca.Sources[0]))
		tgt := b.AddNode(graph.NodeID(ca.Targets[0]))
		b.ConnectNodes(src, tgt)
	}
	return b.Build()
}
