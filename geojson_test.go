package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoadGeoJSON(t *testing.T) {
	collection := GeoJSONFeatureCollection{
		Type: "FeatureCollection",
		Features: []Feature{
			{
				Type:     "Feature",
				Geometry: Geometry{Type: "Point"},
				Properties: Properties{
					ID:            "test-bus",
					Name:          "test-bus",
					AssetType:     "Bus",
					GlossaryTerms: []string{"POWERFLOW"},
					ConnectedAssets: ConnectedAssets{
						Sources: []string{},
						Targets: []string{},
					},
				},
			},
		},
	}

	data, _ := json.Marshal(collection)
	f, err := os.CreateTemp("", "geojson-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write(data)
	f.Close()

	loaded := LoadGeoJSON(f.Name())
	if loaded.Type != "FeatureCollection" {
		t.Errorf("type = %q, want FeatureCollection", loaded.Type)
	}
	if len(loaded.Features) != 1 {
		t.Fatalf("feature count = %d, want 1", len(loaded.Features))
	}
	if loaded.Features[0].Properties.ID != "test-bus" {
		t.Errorf("id = %q, want test-bus", loaded.Features[0].Properties.ID)
	}
}
