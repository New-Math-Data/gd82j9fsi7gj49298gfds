package opendss

import (
	"bytes"
	"encoding/csv"
	"os"
	"strings"

	"opendss-assessment/geojson"
)

// Specs is just an arbitrary jsonserializable JSON object.
// We need it to project arbitrary OpenDSS properties into GeoJSON properties.
type Specs map[string]interface{}

func (s Specs) String(key string) string {
	v, _ := s[key].(string)
	return v
}

// Circuit is the overall model for what's in the OpenDSS file.
type Circuit struct {
	lines    []*Line
	vsources []*Vsource
	buses    map[BusID]*Bus
}

func NewCircuit() *Circuit {
	return &Circuit{buses: make(map[BusID]*Bus)}
}

// Loads a csv file decorating busses with lat/lon.
func (c *Circuit) LoadBusCoords(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		return err
	}
	for _, row := range records {
		if b := NewBusFromCSV(row); b != nil {
			c.buses[b.ID] = b
		}
	}
	return nil
}

// This is the main parser for the OpenDSS master file.
func (c *Circuit) LoadCircuitModel(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Splitting for both windows and unix newlines.
	for _, rawLine := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		// Only interested in Lines and VSources.
		if l := NewLineFromOpenDSS(rawLine); l != nil {
			c.lines = append(c.lines, l)
		} else if v := NewVsourceFromOpenDSS(rawLine); v != nil {
			c.vsources = append(c.vsources, v)
		}
	}
	return nil
}

// Converts the whole Circuit to a giant GeoJSON.
func (c *Circuit) ToGeoJSON() *geojson.FeatureCollection {
	// Convert the buses into their full form in preparation for
	// appending the lines and vsources connected to them. Storing
	// this in a map allows us O(1) lookups as we iterate over the
	// lines and vsources. We keep a copy in an array because we
	// need to flatten this all at the end to write to the GeoJSON.
	busFeatures := make([]*geojson.Feature, 0, len(c.buses))
	busFeatureMap := make(map[BusID]*geojson.Feature, len(c.buses))
	for _, b := range c.buses {
		busFeatures = append(busFeatures, b.ToFeature())
		busFeatureMap[b.ID] = busFeatures[len(busFeatures)-1]
	}

	// Enrich the lines with their connected busses.
	// Because the busses are in a map, this makes the lookups O(1)
	// and we are iterating linearly over the lines, so this remains O(n).
	lineFeatures := make([]*geojson.Feature, len(c.lines))
	for i, l := range c.lines {
		lineFeatures[i] = l.ToFeature(c.buses)
		bus1ID := BusID(lineFeatures[i].Properties.ConnectedAssets.Sources[0])
		bus2ID := BusID(lineFeatures[i].Properties.ConnectedAssets.Targets[0])
		lineID := lineFeatures[i].Properties.ID
		if bf, ok := busFeatureMap[bus1ID]; ok {
			bf.Properties.ConnectedAssets.Targets = append(bf.Properties.ConnectedAssets.Targets, lineID)
		}
		if bf, ok := busFeatureMap[bus2ID]; ok {
			bf.Properties.ConnectedAssets.Sources = append(bf.Properties.ConnectedAssets.Sources, lineID)
		}
	}

	// Enrich the vsources with their busses.
	// Like with lines, this is O(n) because we iterate over the vsources once and use
	// a hashmap with O(1) lookups to get the bus.
	vsourceFeatures := make([]*geojson.Feature, len(c.vsources))
	for i, v := range c.vsources {
		vsourceFeatures[i] = v.ToFeature(c.buses)
		bus1ID := BusID(vsourceFeatures[i].Properties.ConnectedAssets.Targets[0])
		vsrcID := vsourceFeatures[i].Properties.ID
		if bf, ok := busFeatureMap[bus1ID]; ok {
			bf.Properties.ConnectedAssets.Sources = append(bf.Properties.ConnectedAssets.Sources, vsrcID)
		}
	}

	// ...and the geojson file is just a giant assemblage of all the individial
	// elements (busses, lines, and vsources) enriched with their connectivity.
	var features []*geojson.Feature
	features = append(features, busFeatures...)
	features = append(features, lineFeatures...)
	features = append(features, vsourceFeatures...)

	return geojson.NewFeatureCollection(features)
}
