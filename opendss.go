package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func SplitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		line := s[start:]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		lines = append(lines, line)
	}
	return lines
}

func normalizeBusID(raw string) string {
	dotParts := strings.SplitN(raw, ".", 2)
	base := dotParts[0]
	underParts := strings.SplitN(base, "_", 2)
	if len(underParts) > 1 {
		if _, err := strconv.Atoi(underParts[0]); err == nil {
			return underParts[0]
		}
	}
	return base
}

func parseSpecs(tokens []string) map[string]interface{} {
	specs := make(map[string]interface{})
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

func ParseBusCoords(filePath string) ([]Feature, map[string]*Feature) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading bus coords: %v\n", err)
		os.Exit(1)
	}

	var buses []Feature
	for _, line := range SplitLines(string(data)) {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		id := strings.TrimSpace(parts[0])
		lat, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		lon, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)

		coords, _ := json.Marshal([]float64{lon, lat})
		buses = append(buses, Feature{
			Type: "Feature",
			Geometry: Geometry{
				Type:        "Point",
				Coordinates: coords,
			},
			Properties: Properties{
				ID:            id,
				Name:          id,
				AssetType:     "Bus",
				GlossaryTerms: []string{"POWERFLOW"},
				ConnectedAssets: ConnectedAssets{
					Sources: []string{},
					Targets: []string{},
				},
			},
		})
	}

	busMap := make(map[string]*Feature)
	for i := range buses {
		busMap[buses[i].Properties.ID] = &buses[i]
	}
	return buses, busMap
}

func ParseCircuitModel(filePath string) ([]Feature, []Feature) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading circuit model: %v\n", err)
		os.Exit(1)
	}

	var lines []Feature
	var vsources []Feature

	for _, rawLine := range SplitLines(string(data)) {
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
			lines = append(lines, Feature{
				Type: "Feature",
				Geometry: Geometry{
					Type: "LineString",
				},
				Properties: Properties{
					ID:              "Line." + id,
					Name:            "Line." + id,
					AssetType:       "Line",
					Specifications:  specs,
					GlossaryTerms:   []string{},
					ConnectedAssets: ConnectedAssets{Sources: []string{}, Targets: []string{}},
				},
			})
		} else {
			vsources = append(vsources, Feature{
				Type: "Feature",
				Geometry: Geometry{
					Type: "Point",
				},
				Properties: Properties{
					ID:              "Vsource." + id,
					Name:            "Vsource." + id,
					AssetType:       "Vsource",
					Specifications:  specs,
					GlossaryTerms:   []string{"POWERFLOW"},
					ConnectedAssets: ConnectedAssets{Sources: []string{}, Targets: []string{}},
				},
			})
		}
	}

	return lines, vsources
}

func WireConnectivity(buses []Feature, busMap map[string]*Feature, lines []Feature, vsources []Feature) ([]Feature, []Feature) {
	for i := range lines {
		specs := lines[i].Properties.Specifications
		bus1Raw, _ := specs["bus1"].(string)
		bus2Raw, _ := specs["bus2"].(string)
		bus1ID := normalizeBusID(bus1Raw)
		bus2ID := normalizeBusID(bus2Raw)

		lines[i].Properties.ConnectedAssets.Sources = []string{bus1ID}
		lines[i].Properties.ConnectedAssets.Targets = []string{bus2ID}

		b1, ok1 := busMap[bus1ID]
		b2, ok2 := busMap[bus2ID]

		if ok1 && ok2 {
			var coord1, coord2 []float64
			json.Unmarshal(b1.Geometry.Coordinates, &coord1)
			json.Unmarshal(b2.Geometry.Coordinates, &coord2)
			lineCoords, _ := json.Marshal([][]float64{coord1, coord2})
			lines[i].Geometry.Coordinates = lineCoords
		}

		lineRef := lines[i].Properties.ID
		if ok1 {
			b1.Properties.ConnectedAssets.Targets = append(b1.Properties.ConnectedAssets.Targets, lineRef)
		}
		if ok2 {
			b2.Properties.ConnectedAssets.Sources = append(b2.Properties.ConnectedAssets.Sources, lineRef)
		}
	}

	for i := range vsources {
		specs := vsources[i].Properties.Specifications
		bus1Raw, _ := specs["bus1"].(string)
		bus1ID := normalizeBusID(bus1Raw)

		vsources[i].Properties.ConnectedAssets.Targets = []string{bus1ID}
		vsources[i].Properties.ConnectedAssets.Sources = []string{}

		if b, ok := busMap[bus1ID]; ok {
			vsources[i].Geometry.Coordinates = b.Geometry.Coordinates
			b.Properties.ConnectedAssets.Sources = append(b.Properties.ConnectedAssets.Sources, vsources[i].Properties.ID)
		}
	}

	return lines, vsources
}
