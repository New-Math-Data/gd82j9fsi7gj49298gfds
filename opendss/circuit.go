package opendss

import (
	"bytes"
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"opendss-assessment/geojson"
)

type Specs map[string]interface{}

func (s Specs) String(key string) string {
	v, _ := s[key].(string)
	return v
}

type Circuit struct {
	lines    []*Line
	vsources []*Vsource
	busIndex map[BusID]*Bus
}

func NewCircuit() *Circuit {
	return &Circuit{busIndex: make(map[BusID]*Bus)}
}

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
		if len(row) < 3 {
			continue
		}
		id := BusID(strings.TrimSpace(row[0]))
		lat, _ := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		lon, _ := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
		c.busIndex[id] = &Bus{ID: id, Lat: lat, Lon: lon}
	}
	return nil
}

func (c *Circuit) LoadCircuitModel(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
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
	return nil
}

func (c *Circuit) ToGeoJSON() *geojson.GeoJSONFeatureCollection {
	busFeatures := make([]*geojson.Feature, 0, len(c.busIndex))
	busFeatureMap := make(map[BusID]*geojson.Feature, len(c.busIndex))
	for _, b := range c.busIndex {
		busFeatures = append(busFeatures, b.ToFeature())
		busFeatureMap[b.ID] = busFeatures[len(busFeatures)-1]
	}

	lineFeatures := make([]*geojson.Feature, len(c.lines))
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

	vsourceFeatures := make([]*geojson.Feature, len(c.vsources))
	for i, v := range c.vsources {
		vsourceFeatures[i] = v.ToFeature(c.busIndex)
		bus1ID := BusID(vsourceFeatures[i].Properties.ConnectedAssets.Targets[0])
		vsrcID := vsourceFeatures[i].Properties.ID
		if bf, ok := busFeatureMap[bus1ID]; ok {
			bf.Properties.ConnectedAssets.Sources = append(bf.Properties.ConnectedAssets.Sources, vsrcID)
		}
	}

	var features []*geojson.Feature
	features = append(features, busFeatures...)
	features = append(features, lineFeatures...)
	features = append(features, vsourceFeatures...)

	return geojson.NewGeoJSONFeatureCollection(features)
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
