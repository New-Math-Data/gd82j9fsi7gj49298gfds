package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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

// SplitLines splits a string into lines, handling both \n and \r\n line endings.
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

// --- Part 1: Convert Command Functions ---

// ParseBusCoords reads the bus coordinates CSV file and returns a slice of
// Bus features and a map of bus ID -> *Feature for coordinate lookups.
func ParseBusCoords(filePath string) ([]Feature, map[string]*Feature) {
	// TODO: Implement bus coordinate parsing
	// 1. Read the file
	// 2. For each line, split on comma to get: bus_id, latitude, longitude
	// 3. Create a Feature with Point geometry (remember: GeoJSON is [lon, lat])
	// 4. Return the features and a lookup map keyed by bus ID
	return nil, nil
}

// ParseCircuitModel reads the DSS circuit model file and returns Line features
// and Vsource features.
func ParseCircuitModel(filePath string) ([]Feature, []Feature) {
	// TODO: Implement circuit model parsing
	// 1. Read the file and split into lines
	// 2. Identify Line definitions (start with: New "Line.<id>")
	// 3. Identify Vsource definitions (second field contains "Vsource.<id>", can be New or Edit)
	// 4. For each Line, parse key=value pairs into Specifications
	// 5. For each Vsource, parse key=value pairs into Specifications
	// 6. Return the two slices of features
	return nil, nil
}

// WireConnectivity connects Lines and Vsources to their respective Buses,
// resolving geometries and populating connected_assets on all features.
func WireConnectivity(buses []Feature, busMap map[string]*Feature, lines []Feature, vsources []Feature) ([]Feature, []Feature) {
	// TODO: Implement connectivity wiring
	// 1. For each Line: normalize bus1/bus2 IDs, set connected_assets, resolve LineString geometry
	// 2. For each Vsource: set connected_assets.targets to the connected bus
	// 3. Update Bus connected_assets to reference their connected Lines and Vsources
	// 4. Resolve Vsource geometry from the connected bus coordinates
	return lines, vsources
}

// RunConvert executes the "convert" subcommand: parses OpenDSS data and writes
// a GeoJSON FeatureCollection to the specified output file.
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

// LoadGeoJSON reads a GeoJSON FeatureCollection from a file and returns
// the parsed collection.
func LoadGeoJSON(filePath string) GeoJSONFeatureCollection {
	// TODO: Implement GeoJSON loading
	// 1. Read the file
	// 2. Unmarshal JSON into a GeoJSONFeatureCollection
	// 3. Return the collection
	return GeoJSONFeatureCollection{}
}

// BuildGraph constructs an adjacency list representation of the network graph
// from a GeoJSON FeatureCollection. Buses are nodes, Lines are undirected edges.
func BuildGraph(collection GeoJSONFeatureCollection) map[string][]string {
	// TODO: Implement graph construction
	// 1. Iterate over features, filtering for assetType == "Line"
	// 2. For each Line, read connected_assets.sources[0] and connected_assets.targets[0]
	// 3. Add an undirected edge between them in the adjacency list
	// 4. Return the adjacency list as map[busID] -> []neighborBusIDs
	return nil
}

// ShortestPath computes the shortest distance (hop count) between two nodes
// in an unweighted graph. Returns -1 if no path exists.
func ShortestPath(graph map[string][]string, source, target string) int {
	// TODO: Implement BFS shortest path
	// 1. Use breadth-first search starting from source
	// 2. Track visited nodes and current depth
	// 3. Return the depth when target is found
	// 4. Return -1 if the queue is exhausted without finding target
	return -1
}

// RunDistance executes the "distance" subcommand: loads a GeoJSON file,
// builds a graph, and computes the shortest path between two bus IDs.
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
