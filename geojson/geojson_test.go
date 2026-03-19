package geojson

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestLoadGeoJSON(t *testing.T) {
	// Tests loading one of every type of object: Line, Bus, Vsource
	marshalled := &GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		Features: []Feature{
			{
				Type:     "Feature",
				Geometry: Geometry{Type: "Point", Coordinates: json.RawMessage(`[-85.1481,35.05225]`)},
				Properties: Properties{
					ID:            "tyn201",
					Name:          "tyn201",
					AssetType:     "Bus",
					GlossaryTerms: []string{"POWERFLOW"},
					ConnectedAssets: ConnectedAssets{
						Sources: []string{"Vsource.source"},
						Targets: []string{"Line.tyn201_feeder"},
					},
				},
			},
			{
				Type:     "Feature",
				Geometry: Geometry{Type: "LineString", Coordinates: json.RawMessage(`[[-85.1481,35.05225],[-85.14808268,35.05217897]]`)},
				Properties: Properties{
					ID:        "Line.tyn201_feeder",
					Name:      "Line.tyn201_feeder",
					AssetType: "Line",
					Specifications: map[string]interface{}{
						"phases":    "3",
						"bus1":      "tyn201.1.2.3",
						"bus2":      "1147583.1.2.3",
						"length":    "30.7848",
						"units":     "m",
						"spacing":   "3PH_HORIZ_LG_1C_3",
						"wires":     "\"AL_556_19STR AL_556_19STR AL_556_19STR ACSR_1/0_6/1\"",
						"Seasons":   "1",
						"Ratings":   "[400,]",
						"normamps":  "613",
						"emergamps": "613",
					},
					GlossaryTerms: []string{},
					ConnectedAssets: ConnectedAssets{
						Sources: []string{"tyn201"},
						Targets: []string{"1147583"},
					},
				},
			},
			{
				Type:     "Feature",
				Geometry: Geometry{Type: "Point", Coordinates: json.RawMessage(`[-85.1481,35.05225]`)},
				Properties: Properties{
					ID:        "Vsource.source",
					Name:      "Vsource.source",
					AssetType: "Vsource",
					Specifications: map[string]interface{}{
						"bus1":   "tyn201",
						"basekv": "12.47",
						"pu":     "1.00000001989",
						"angle":  "0.000000",
						"Z1":     "[0.141506, 1.399508]",
						"Z0":     "[0.054425, 1.088506]",
					},
					GlossaryTerms: []string{"POWERFLOW"},
					ConnectedAssets: ConnectedAssets{
						Sources: []string{},
						Targets: []string{"tyn201"},
					},
				},
			},
		},
	}

	raw := `
{
	"type": "FeatureCollection",
	"features": [
		{
			"type": "Feature",
			"geometry": {
				"type": "Point",
				"coordinates": [
					-85.1481,
					35.05225
				]
			},
			"properties": {
				"id": "tyn201",
				"name": "tyn201",
				"assetType": "Bus",
				"connected_assets": {
					"sources": [
						"Vsource.source"
					],
					"targets": [
						"Line.tyn201_feeder"
					]
				},
				"glossary_terms": [
					"POWERFLOW"
				]
			}
		},
		{
			"type": "Feature",
			"geometry": {
				"type": "LineString",
				"coordinates": [
					[
						-85.1481,
						35.05225
					],
					[
						-85.14808268,
						35.05217897
					]
				]
			},
			"properties": {
				"id": "Line.tyn201_feeder",
				"name": "Line.tyn201_feeder",
				"assetType": "Line",
				"specifications": {
					"phases": "3",
					"bus1": "tyn201.1.2.3",
					"bus2": "1147583.1.2.3",
					"length": "30.7848",
					"units": "m",
					"spacing": "3PH_HORIZ_LG_1C_3",
					"wires": "\"AL_556_19STR AL_556_19STR AL_556_19STR ACSR_1/0_6/1\"",
					"Seasons": "1",
					"Ratings": "[400,]",
					"normamps": "613",
					"emergamps": "613"
				},
				"connected_assets": {
					"sources": [
						"tyn201"
					],
					"targets": [
						"1147583"
					]
				},
				"glossary_terms": []
			}
		},
		{
			"type": "Feature",
			"geometry": {
				"type": "Point",
				"coordinates": [
					-85.1481,
					35.05225
				]
			},
			"properties": {
				"id": "Vsource.source",
				"name": "Vsource.source",
				"assetType": "Vsource",
				"specifications": {
					"bus1": "tyn201",
					"basekv": "12.47",
					"pu": "1.00000001989",
					"angle": "0.000000",
					"Z1": "[0.141506, 1.399508]",
					"Z0": "[0.054425, 1.088506]"
				},
				"connected_assets": {
					"sources": [],
					"targets": [
						"tyn201"
					]
				},
				"glossary_terms": [
					"POWERFLOW"
				]
			}
		}
	]
}
`

	f, err := os.CreateTemp("", "geojson-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write([]byte(raw))
	f.Close()

	loaded, err := LoadGeoJSON(f.Name())
	assert.NoError(t, err)

	wantJSON, err := json.Marshal(marshalled)
	assert.NoError(t, err)

	gotJSON, err := json.Marshal(loaded)
	assert.NoError(t, err)

	assert.JSONEq(t, string(wantJSON), string(gotJSON))
}

func TestLoadGeoJSONFileNotExists(t *testing.T) {
	loaded, err := LoadGeoJSON("nonexistent.json")
	assert.Error(t, err)
	assert.Nil(t, loaded)
}

func TestLoadGeoJSONEmptyFile(t *testing.T) {
	f, err := os.CreateTemp("", "geojson-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	loaded, err := LoadGeoJSON(f.Name())
	assert.Error(t, err)
	assert.Nil(t, loaded)
}

func TestLoadGeoJSONUnparseable(t *testing.T) {
	f, err := os.CreateTemp("", "geojson-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write([]byte(`{not valid json`))
	f.Close()

	loaded, err := LoadGeoJSON(f.Name())
	assert.Error(t, err)
	assert.Nil(t, loaded)
}

func TestLoadGeoJSONEmptyFeatures(t *testing.T) {
	f, err := os.CreateTemp("", "geojson-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write([]byte(`{"type":"FeatureCollection","features":[]}`))
	f.Close()

	loaded, err := LoadGeoJSON(f.Name())
	assert.NoError(t, err)
	assert.Empty(t, loaded.Features)
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
	g := collection.BuildGraph()
	foundAB, foundBA := false, false
	for _, e := range g.Edges() {
		if e.From.ID == "A" && e.To.ID == "B" {
			foundAB = true
		}
		if e.From.ID == "B" && e.To.ID == "A" {
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
