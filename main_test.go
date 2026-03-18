package main

import (
	"encoding/json"
	"testing"
)

func TestNormalizeBusID(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"tyn201.1.2.3", "tyn201"},
		{"1001176.3", "1001176"},
		{"108687_1s.1.2.3", "108687"},
		{"tyn201", "tyn201"},
		{"1147844_1s.1.2.3", "1147844"},
	}
	for _, c := range cases {
		got := normalizeBusID(c.input)
		if got != c.want {
			t.Errorf("normalizeBusID(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestShortestPathSameNode(t *testing.T) {
	graph := map[string][]string{
		"A": {"B"},
		"B": {"A"},
	}
	if d := ShortestPath(graph, "A", "A"); d != 0 {
		t.Errorf("expected 0, got %d", d)
	}
}

func TestShortestPathConnected(t *testing.T) {
	graph := map[string][]string{
		"A": {"B"},
		"B": {"A", "C"},
		"C": {"B"},
	}
	if d := ShortestPath(graph, "A", "C"); d != 2 {
		t.Errorf("expected 2, got %d", d)
	}
	if d := ShortestPath(graph, "A", "B"); d != 1 {
		t.Errorf("expected 1, got %d", d)
	}
}

func TestShortestPathDisconnected(t *testing.T) {
	graph := map[string][]string{
		"A": {"B"},
		"B": {"A"},
		"C": {"D"},
		"D": {"C"},
	}
	if d := ShortestPath(graph, "A", "C"); d != -1 {
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
	graph := BuildGraph(collection)
	found := false
	for _, n := range graph["A"] {
		if n == "B" {
			found = true
		}
	}
	if !found {
		t.Error("expected A -> B edge")
	}
	found = false
	for _, n := range graph["B"] {
		if n == "A" {
			found = true
		}
	}
	if !found {
		t.Error("expected B -> A edge (undirected)")
	}
}

func TestParseSpecsQuotedMultiToken(t *testing.T) {
	tokens := []string{`wires="AL_556_19STR`, `AL_556_19STR`, `AL_556_19STR`, `ACSR_1/0_6/1"`, `normamps=613`}
	specs := parseSpecs(tokens)
	want := `"AL_556_19STR AL_556_19STR AL_556_19STR ACSR_1/0_6/1"`
	if got, _ := specs["wires"].(string); got != want {
		t.Errorf("wires = %q, want %q", got, want)
	}
	if got, _ := specs["normamps"].(string); got != "613" {
		t.Errorf("normamps = %q, want %q", got, "613")
	}
}

func TestParseSpecsBracketArrayTokens(t *testing.T) {
	tokens := []string{`Z1=[0.141506,`, `1.399508]`}
	specs := parseSpecs(tokens)
	want := "[0.141506, 1.399508]"
	if got, _ := specs["Z1"].(string); got != want {
		t.Errorf("Z1 = %q, want %q", got, want)
	}
	if _, exists := specs["1.399508]"]; exists {
		t.Error(`specs["1.399508]"] should not exist as a key`)
	}
}

func TestParseBusCoords(t *testing.T) {
	buses, busMap := ParseBusCoords("data/busGISCoords.csv")
	if len(buses) == 0 {
		t.Fatal("expected buses, got none")
	}

	b, ok := busMap["tyn201"]
	if !ok {
		t.Fatal("tyn201 not found in busMap")
	}
	if b.Properties.AssetType != "Bus" {
		t.Errorf("assetType = %q, want Bus", b.Properties.AssetType)
	}
	if len(b.Properties.GlossaryTerms) != 1 || b.Properties.GlossaryTerms[0] != "POWERFLOW" {
		t.Errorf("glossary_terms = %v, want [POWERFLOW]", b.Properties.GlossaryTerms)
	}

	var coords []float64
	if err := json.Unmarshal(b.Geometry.Coordinates, &coords); err != nil {
		t.Fatalf("failed to unmarshal coords: %v", err)
	}
	if len(coords) != 2 || coords[0] != -85.1481 || coords[1] != 35.05225 {
		t.Errorf("coords = %v, want [-85.1481 35.05225]", coords)
	}
}

func TestParseCircuitModel(t *testing.T) {
	lines, vsources := ParseCircuitModel("data/master.dss")
	if len(lines) == 0 {
		t.Fatal("expected lines, got none")
	}
	if len(vsources) == 0 {
		t.Fatal("expected vsources, got none")
	}

	var feeder *Feature
	for i := range lines {
		if lines[i].Properties.ID == "Line.tyn201_feeder" {
			feeder = &lines[i]
			break
		}
	}
	if feeder == nil {
		t.Fatal("Line.tyn201_feeder not found")
	}
	if bus1, _ := feeder.Properties.Specifications["bus1"].(string); bus1 != "tyn201.1.2.3" {
		t.Errorf("bus1 = %q, want tyn201.1.2.3", bus1)
	}

	var vsrc *Feature
	for i := range vsources {
		if vsources[i].Properties.ID == "Vsource.source" {
			vsrc = &vsources[i]
			break
		}
	}
	if vsrc == nil {
		t.Fatal("Vsource.source not found")
	}
}

func TestWireConnectivity(t *testing.T) {
	buses, busMap := ParseBusCoords("data/busGISCoords.csv")
	lines, vsources := ParseCircuitModel("data/master.dss")
	lines, vsources = WireConnectivity(buses, busMap, lines, vsources)

	var feeder *Feature
	for i := range lines {
		if lines[i].Properties.ID == "Line.tyn201_feeder" {
			feeder = &lines[i]
			break
		}
	}
	if feeder == nil {
		t.Fatal("Line.tyn201_feeder not found after wiring")
	}
	if len(feeder.Properties.ConnectedAssets.Sources) != 1 || feeder.Properties.ConnectedAssets.Sources[0] != "tyn201" {
		t.Errorf("feeder sources = %v, want [tyn201]", feeder.Properties.ConnectedAssets.Sources)
	}
	if len(feeder.Properties.ConnectedAssets.Targets) != 1 || feeder.Properties.ConnectedAssets.Targets[0] != "1147583" {
		t.Errorf("feeder targets = %v, want [1147583]", feeder.Properties.ConnectedAssets.Targets)
	}

	var vsrc *Feature
	for i := range vsources {
		if vsources[i].Properties.ID == "Vsource.source" {
			vsrc = &vsources[i]
			break
		}
	}
	if vsrc == nil {
		t.Fatal("Vsource.source not found after wiring")
	}
	if len(vsrc.Properties.ConnectedAssets.Targets) != 1 || vsrc.Properties.ConnectedAssets.Targets[0] != "tyn201" {
		t.Errorf("vsource targets = %v, want [tyn201]", vsrc.Properties.ConnectedAssets.Targets)
	}

	tyn201 := busMap["tyn201"]
	foundVsrc := false
	for _, s := range tyn201.Properties.ConnectedAssets.Sources {
		if s == "Vsource.source" {
			foundVsrc = true
		}
	}
	if !foundVsrc {
		t.Errorf("tyn201 sources = %v, want to contain Vsource.source", tyn201.Properties.ConnectedAssets.Sources)
	}
}
