package main

import "testing"

func TestNewBusID(t *testing.T) {
	cases := []struct {
		input string
		want  BusID
	}{
		{"tyn201.1.2.3", "tyn201"},
		{"1001176.3", "1001176"},
		{"108687_1s.1.2.3", "108687"},
		{"tyn201", "tyn201"},
		{"1147844_1s.1.2.3", "1147844"},
	}
	for _, c := range cases {
		got := NewBusID(c.input)
		if got != c.want {
			t.Errorf("NewBusID(%q) = %q, want %q", c.input, got, c.want)
		}
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

func TestLoadBusCoords(t *testing.T) {
	c := NewCircuit()
	if err := c.LoadBusCoords("data/busGISCoords.csv"); err != nil {
		t.Fatal(err)
	}

	if len(c.buses) == 0 {
		t.Fatal("expected buses, got none")
	}

	b, ok := c.busIndex[BusID("tyn201")]
	if !ok {
		t.Fatal("tyn201 not found in busIndex")
	}
	if b.Lat != 35.05225 || b.Lon != -85.1481 {
		t.Errorf("coords = (%v, %v), want (35.05225, -85.1481)", b.Lat, b.Lon)
	}
}

func TestLoadCircuitModel(t *testing.T) {
	c := NewCircuit()
	if err := c.LoadCircuitModel("data/master.dss"); err != nil {
		t.Fatal(err)
	}

	if len(c.lines) == 0 {
		t.Fatal("expected lines, got none")
	}
	if len(c.vsources) == 0 {
		t.Fatal("expected vsources, got none")
	}

	var feeder *Line
	for _, l := range c.lines {
		if l.ID == "tyn201_feeder" {
			feeder = l
			break
		}
	}
	if feeder == nil {
		t.Fatal("tyn201_feeder not found")
	}
	if bus1 := feeder.Specs.String("bus1"); bus1 != "tyn201.1.2.3" {
		t.Errorf("bus1 = %q, want tyn201.1.2.3", bus1)
	}

	var vsrc *Vsource
	for _, v := range c.vsources {
		if v.ID == "source" {
			vsrc = v
			break
		}
	}
	if vsrc == nil {
		t.Fatal("Vsource.source not found")
	}
}

func TestToGeoJSON(t *testing.T) {
	c := NewCircuit()
	if err := c.LoadBusCoords("data/busGISCoords.csv"); err != nil {
		t.Fatal(err)
	}
	if err := c.LoadCircuitModel("data/master.dss"); err != nil {
		t.Fatal(err)
	}
	collection := c.ToGeoJSON()

	var feeder, tyn201Bus, vsrcFeature *Feature
	for i := range collection.Features {
		switch collection.Features[i].Properties.ID {
		case "Line.tyn201_feeder":
			feeder = &collection.Features[i]
		case "tyn201":
			tyn201Bus = &collection.Features[i]
		case "Vsource.source":
			vsrcFeature = &collection.Features[i]
		}
	}

	if feeder == nil {
		t.Fatal("Line.tyn201_feeder not found")
	}
	if len(feeder.Properties.ConnectedAssets.Sources) != 1 || feeder.Properties.ConnectedAssets.Sources[0] != "tyn201" {
		t.Errorf("feeder sources = %v, want [tyn201]", feeder.Properties.ConnectedAssets.Sources)
	}
	if len(feeder.Properties.ConnectedAssets.Targets) != 1 || feeder.Properties.ConnectedAssets.Targets[0] != "1147583" {
		t.Errorf("feeder targets = %v, want [1147583]", feeder.Properties.ConnectedAssets.Targets)
	}

	if vsrcFeature == nil {
		t.Fatal("Vsource.source not found")
	}
	if len(vsrcFeature.Properties.ConnectedAssets.Targets) != 1 || vsrcFeature.Properties.ConnectedAssets.Targets[0] != "tyn201" {
		t.Errorf("vsource targets = %v, want [tyn201]", vsrcFeature.Properties.ConnectedAssets.Targets)
	}

	if tyn201Bus == nil {
		t.Fatal("tyn201 bus not found")
	}
	foundVsrc := false
	for _, s := range tyn201Bus.Properties.ConnectedAssets.Sources {
		if s == "Vsource.source" {
			foundVsrc = true
		}
	}
	if !foundVsrc {
		t.Errorf("tyn201 sources = %v, want to contain Vsource.source", tyn201Bus.Properties.ConnectedAssets.Sources)
	}
}
