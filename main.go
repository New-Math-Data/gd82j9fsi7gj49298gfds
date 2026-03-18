package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// --- GeoJSON Type Definitions ---

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

// --- Utility Functions ---

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

// --- Part 1: Convert Command Functions ---

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

func parseSpecs(tokens []string) map[string]interface{} {
	specs := make(map[string]interface{})
	var accumKey string
	var accumParts []string

	for _, tok := range tokens {
		if accumKey != "" {
			accumParts = append(accumParts, tok)
			if strings.HasSuffix(tok, "\"") {
				specs[accumKey] = strings.Join(accumParts, " ")
				accumKey = ""
				accumParts = nil
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
			} else {
				specs[k] = v
			}
		} else {
			specs[tok] = "true"
		}
	}
	return specs
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

func RunConvert(circuitModel, busCoords, outputFile string) {
	buses, busMap := ParseBusCoords(busCoords)
	lines, vsources := ParseCircuitModel(circuitModel)
	lines, vsources = WireConnectivity(buses, busMap, lines, vsources)

	var features []Feature
	features = append(features, buses...)
	features = append(features, lines...)
	features = append(features, vsources...)

	collection := GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

	output, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling GeoJSON: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wrote %s\n", outputFile)
}

// --- Part 2: Distance Command Functions ---

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

func BuildGraph(collection GeoJSONFeatureCollection) map[string][]string {
	graph := make(map[string][]string)
	for _, f := range collection.Features {
		if f.Properties.AssetType != "Line" {
			continue
		}
		ca := f.Properties.ConnectedAssets
		if len(ca.Sources) == 0 || len(ca.Targets) == 0 {
			continue
		}
		bus1 := ca.Sources[0]
		bus2 := ca.Targets[0]
		graph[bus1] = append(graph[bus1], bus2)
		graph[bus2] = append(graph[bus2], bus1)
	}
	return graph
}

func ShortestPath(graph map[string][]string, source, target string) int {
	if source == target {
		return 0
	}
	type item struct {
		node string
		dist int
	}
	visited := map[string]bool{source: true}
	queue := []item{{source, 0}}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, neighbor := range graph[curr.node] {
			if neighbor == target {
				return curr.dist + 1
			}
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, item{neighbor, curr.dist + 1})
			}
		}
	}
	return -1
}

func RunDistance(inputFile, sourceNode, targetNode string) {
	collection := LoadGeoJSON(inputFile)
	graph := BuildGraph(collection)
	distance := ShortestPath(graph, sourceNode, targetNode)
	fmt.Printf("Shortest path from %s to %s: %d\n", sourceNode, targetNode, distance)
}

// --- Main ---

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  convert   Convert OpenDSS data to GeoJSON\n")
		fmt.Fprintf(os.Stderr, "  distance  Compute shortest path between two buses\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "convert":
		convertCmd := flag.NewFlagSet("convert", flag.ExitOnError)
		circuitModel := convertCmd.String("circuit_model", "", "Path to the circuit model file (required)")
		busCoords := convertCmd.String("bus_coords", "", "Path to the bus coordinates file (required)")
		outputFile := convertCmd.String("output", "output.json", "Path to the output GeoJSON file")
		convertCmd.Parse(os.Args[2:])

		if *circuitModel == "" || *busCoords == "" {
			fmt.Fprintf(os.Stderr, "Error: -circuit_model and -bus_coords are required\n\n")
			convertCmd.Usage()
			os.Exit(1)
		}

		RunConvert(*circuitModel, *busCoords, *outputFile)

	case "distance":
		distanceCmd := flag.NewFlagSet("distance", flag.ExitOnError)
		inputFile := distanceCmd.String("input", "output.json", "Path to the GeoJSON file to read")
		sourceNode := distanceCmd.String("source_node", "", "Source bus ID (required)")
		targetNode := distanceCmd.String("target_node", "", "Target bus ID (required)")
		distanceCmd.Parse(os.Args[2:])

		if *sourceNode == "" || *targetNode == "" {
			fmt.Fprintf(os.Stderr, "Error: -source_node and -target_node are required\n\n")
			distanceCmd.Usage()
			os.Exit(1)
		}

		RunDistance(*inputFile, *sourceNode, *targetNode)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  convert   Convert OpenDSS data to GeoJSON\n")
		fmt.Fprintf(os.Stderr, "  distance  Compute shortest path between two buses\n")
		os.Exit(1)
	}
}
