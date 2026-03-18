package main

import "testing"

func TestShortestPathSameNode(t *testing.T) {
	g := Graph{edges: map[string][]string{
		"A": {"B"},
		"B": {"A"},
	}}
	if d := g.ShortestPath("A", "A"); d != 0 {
		t.Errorf("expected 0, got %d", d)
	}
}

func TestShortestPathConnected(t *testing.T) {
	g := Graph{edges: map[string][]string{
		"A": {"B"},
		"B": {"A", "C"},
		"C": {"B"},
	}}
	if d := g.ShortestPath("A", "C"); d != 2 {
		t.Errorf("expected 2, got %d", d)
	}
	if d := g.ShortestPath("A", "B"); d != 1 {
		t.Errorf("expected 1, got %d", d)
	}
}

func TestShortestPathDisconnected(t *testing.T) {
	g := Graph{edges: map[string][]string{
		"A": {"B"},
		"B": {"A"},
		"C": {"D"},
		"D": {"C"},
	}}
	if d := g.ShortestPath("A", "C"); d != -1 {
		t.Errorf("expected -1, got %d", d)
	}
}

func TestBuildGraphUndirected(t *testing.T) {
	collection := GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		Features: []Feature{
			{
				Type:     "Feature",
				Geometry: Geometry{Type: "LineString"},
				Properties: Properties{
					ID:            "Line.ab",
					AssetType:     "Line",
					GlossaryTerms: []string{},
					ConnectedAssets: ConnectedAssets{
						Sources: []string{"A"},
						Targets: []string{"B"},
					},
				},
			},
		},
	}
	g := BuildGraph(collection)
	foundAB, foundBA := false, false
	for _, n := range g.edges["A"] {
		if n == "B" {
			foundAB = true
		}
	}
	for _, n := range g.edges["B"] {
		if n == "A" {
			foundBA = true
		}
	}
	if !foundAB {
		t.Error("expected A -> B edge")
	}
	if !foundBA {
		t.Error("expected B -> A edge (undirected)")
	}
}
