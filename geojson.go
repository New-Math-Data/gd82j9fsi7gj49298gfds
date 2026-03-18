package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type GeoJSONFeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string     `json:"type"`
	Geometry   Geometry   `json:"geometry"`
	Properties Properties `json:"properties"`
}

type Geometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

type ConnectedAssets struct {
	Sources []string `json:"sources"`
	Targets []string `json:"targets"`
}

type Properties struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	AssetType       string                 `json:"assetType"`
	Specifications  map[string]interface{} `json:"specifications,omitempty"`
	ConnectedAssets ConnectedAssets        `json:"connected_assets"`
	GlossaryTerms   []string               `json:"glossary_terms"`
}

func LoadGeoJSON(filePath string) GeoJSONFeatureCollection {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading GeoJSON: %v\n", err)
		os.Exit(1)
	}
	var collection GeoJSONFeatureCollection
	json.Unmarshal(data, &collection)
	return collection
}
