package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type BusID string

var busIDRe = regexp.MustCompile(`^(\d+)[_.]|^([^.]+)`)

func NewBusID(raw string) BusID {
	m := busIDRe.FindStringSubmatch(raw)
	if m == nil {
		return BusID(raw)
	}
	if m[1] != "" {
		return BusID(m[1])
	}
	return BusID(m[2])
}

type Specs map[string]interface{}

func (s Specs) String(key string) string {
	v, _ := s[key].(string)
	return v
}

type Bus struct {
	ID  BusID
	Lat float64
	Lon float64
}

type Line struct {
	ID    string
	Specs Specs
}

type Vsource struct {
	ID    string
	Specs Specs
}

type Circuit struct {
	buses    []*Bus
	lines    []*Line
	vsources []*Vsource
	busIndex map[BusID]*Bus
}

func NewCircuit() *Circuit {
	return &Circuit{busIndex: make(map[BusID]*Bus)}
}

func (c *Circuit) LoadBusCoords(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading bus coords: %v\n", err)
		os.Exit(1)
	}
	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing bus coords: %v\n", err)
		os.Exit(1)
	}
	for _, row := range records {
		if len(row) < 3 {
			continue
		}
		id := BusID(strings.TrimSpace(row[0]))
		lat, _ := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		lon, _ := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
		b := &Bus{ID: id, Lat: lat, Lon: lon}
		c.buses = append(c.buses, b)
		c.busIndex[id] = b
	}
}

func (c *Circuit) LoadCircuitModel(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading circuit model: %v\n", err)
		os.Exit(1)
	}
	for _, rawLine := range splitLines(string(data)) {
		rawLine = strings.TrimSpace(rawLine)
		if rawLine == "" {
			continue
		}
		tokens := strings.Fields(rawLine)
		if len(tokens) < 2 {
			continue
		}

		isLine := tokens[0] == "New" && strings.HasPrefix(tokens[1], "\"Line.")
		isVsource := strings.Contains(tokens[1], "\"Vsource.")

		if !isLine && !isVsource {
			continue
		}

		fullID := strings.Trim(tokens[1], "\"")
		parts := strings.SplitN(fullID, ".", 2)
		id := ""
		if len(parts) == 2 {
			id = parts[1]
		}

		specs := parseSpecs(tokens[2:])

		if isLine {
			c.lines = append(c.lines, &Line{ID: id, Specs: specs})
		} else {
			c.vsources = append(c.vsources, &Vsource{ID: id, Specs: specs})
		}
	}
}

func (b *Bus) ToFeature() Feature {
	coords, _ := json.Marshal([]float64{b.Lon, b.Lat})
	return Feature{
		Type:     "Feature",
		Geometry: Geometry{Type: "Point", Coordinates: coords},
		Properties: Properties{
			ID:              string(b.ID),
			Name:            string(b.ID),
			AssetType:       "Bus",
			GlossaryTerms:   []string{"POWERFLOW"},
			ConnectedAssets: ConnectedAssets{Sources: []string{}, Targets: []string{}},
		},
	}
}

func (l *Line) ToFeature(busIndex map[BusID]*Bus) Feature {
	bus1ID := NewBusID(l.Specs.String("bus1"))
	bus2ID := NewBusID(l.Specs.String("bus2"))

	var coords json.RawMessage
	b1, ok1 := busIndex[bus1ID]
	b2, ok2 := busIndex[bus2ID]
	if ok1 && ok2 {
		coords, _ = json.Marshal([][]float64{{b1.Lon, b1.Lat}, {b2.Lon, b2.Lat}})
	}

	id := "Line." + l.ID
	return Feature{
		Type:     "Feature",
		Geometry: Geometry{Type: "LineString", Coordinates: coords},
		Properties: Properties{
			ID:              id,
			Name:            id,
			AssetType:       "Line",
			Specifications:  l.Specs,
			GlossaryTerms:   []string{},
			ConnectedAssets: ConnectedAssets{Sources: []string{string(bus1ID)}, Targets: []string{string(bus2ID)}},
		},
	}
}

func (v *Vsource) ToFeature(busIndex map[BusID]*Bus) Feature {
	bus1ID := NewBusID(v.Specs.String("bus1"))

	var coords json.RawMessage
	if b, ok := busIndex[bus1ID]; ok {
		coords, _ = json.Marshal([]float64{b.Lon, b.Lat})
	}

	id := "Vsource." + v.ID
	return Feature{
		Type:     "Feature",
		Geometry: Geometry{Type: "Point", Coordinates: coords},
		Properties: Properties{
			ID:              id,
			Name:            id,
			AssetType:       "Vsource",
			Specifications:  v.Specs,
			GlossaryTerms:   []string{"POWERFLOW"},
			ConnectedAssets: ConnectedAssets{Sources: []string{}, Targets: []string{string(bus1ID)}},
		},
	}
}

func (c *Circuit) ToGeoJSON() GeoJSONFeatureCollection {
	busFeatures := make([]Feature, len(c.buses))
	busFeatureMap := make(map[BusID]*Feature)
	for i, b := range c.buses {
		busFeatures[i] = b.ToFeature()
		busFeatureMap[b.ID] = &busFeatures[i]
	}

	lineFeatures := make([]Feature, len(c.lines))
	for i, l := range c.lines {
		lineFeatures[i] = l.ToFeature(c.busIndex)
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

	vsourceFeatures := make([]Feature, len(c.vsources))
	for i, v := range c.vsources {
		vsourceFeatures[i] = v.ToFeature(c.busIndex)
		bus1ID := BusID(vsourceFeatures[i].Properties.ConnectedAssets.Targets[0])
		vsrcID := vsourceFeatures[i].Properties.ID
		if bf, ok := busFeatureMap[bus1ID]; ok {
			bf.Properties.ConnectedAssets.Sources = append(bf.Properties.ConnectedAssets.Sources, vsrcID)
		}
	}

	var features []Feature
	features = append(features, busFeatures...)
	features = append(features, lineFeatures...)
	features = append(features, vsourceFeatures...)

	return GeoJSONFeatureCollection{Type: "FeatureCollection", Features: features}
}

func splitLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

func parseSpecs(tokens []string) Specs {
	specs := make(Specs)
	var accumKey string
	var accumParts []string
	var accumEnd byte

	for _, tok := range tokens {
		if accumKey != "" {
			accumParts = append(accumParts, tok)
			if len(tok) > 0 && tok[len(tok)-1] == accumEnd {
				specs[accumKey] = strings.Join(accumParts, " ")
				accumKey = ""
				accumParts = nil
				accumEnd = 0
			}
			continue
		}
		eqIdx := strings.Index(tok, "=")
		if eqIdx >= 0 {
			k := tok[:eqIdx]
			v := tok[eqIdx+1:]
			if strings.HasPrefix(v, "\"") && !strings.HasSuffix(v, "\"") {
				accumKey = k
				accumParts = []string{v}
				accumEnd = '"'
			} else if strings.HasPrefix(v, "[") && !strings.HasSuffix(v, "]") {
				accumKey = k
				accumParts = []string{v}
				accumEnd = ']'
			} else {
				specs[k] = v
			}
		} else {
			specs[tok] = "true"
		}
	}
	return specs
}
