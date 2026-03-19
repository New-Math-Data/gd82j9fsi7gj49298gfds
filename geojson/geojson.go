package geojson

import (
	"encoding/json"
	"os"
)

type GeoJSONFeatureCollection struct { //nolint:revive // GeoJSONFeatureCollection is simply the defined name, even though it stutters
	Type     string     `json:"type"`
	Features []*Feature `json:"features"`
}

func NewGeoJSONFeatureCollection(features []*Feature) *GeoJSONFeatureCollection {
	return &GeoJSONFeatureCollection{Type: "FeatureCollection", Features: features}
}

type Feature struct {
	Type       string     `json:"type"`
	Geometry   Geometry   `json:"geometry"`
	Properties Properties `json:"properties"`
}

type Geometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"` // structure varies depending on Point or LineString
}

type ConnectedAssets struct {
	Sources []string `json:"sources"`
	Targets []string `json:"targets"`
}

type Properties struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	AssetType       string                 `json:"assetType"`
	Specifications  map[string]interface{} `json:"specifications,omitempty"` // can be anything jsonserializable as a json object
	ConnectedAssets ConnectedAssets        `json:"connected_assets"`
	GlossaryTerms   []string               `json:"glossary_terms"`
}

func LoadGeoJSON(filePath string) (*GeoJSONFeatureCollection, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	// GeoJSONFeatureCollection is just a giant JSON, so we can unmarshal it directly!
	var collection GeoJSONFeatureCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, err
	}
	return &collection, nil
}
