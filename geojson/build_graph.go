package geojson

import "opendss-assessment/graph"

func (c *GeoJSONFeatureCollection) BuildGraph() *graph.Graph {
	b := graph.NewBuilder()
	for _, f := range c.Features {
		if f.Properties.AssetType != "Line" {
			continue
		}
		ca := f.Properties.ConnectedAssets
		if len(ca.Sources) == 0 || len(ca.Targets) == 0 {
			continue
		}
		src, tgt := graph.NodeID(ca.Sources[0]), graph.NodeID(ca.Targets[0])
		b.AddEdge(src, tgt)
		b.AddEdge(tgt, src)
	}
	return b.Build()
}
