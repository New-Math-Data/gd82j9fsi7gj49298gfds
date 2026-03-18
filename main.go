package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func normalizeDashes(args []string) []string {
	normalized := make([]string, len(args))
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			name := strings.SplitN(arg[1:], "=", 2)[0]
			if len(name) > 1 {
				arg = "-" + arg
			}
		}
		normalized[i] = arg
	}
	return normalized
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

func RunDistance(inputFile, sourceNode, targetNode string) {
	collection := LoadGeoJSON(inputFile)
	graph := BuildGraph(collection)
	distance := ShortestPath(graph, sourceNode, targetNode)
	fmt.Printf("Shortest path from %s to %s: %d\n", sourceNode, targetNode, distance)
}

func main() {
	root := &cobra.Command{
		Use:   "opendss-assessment",
		Short: "OpenDSS to GeoJSON converter and graph path finder",
	}

	var circuitModel, busCoords, outputFile string
	convertCmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert OpenDSS data to GeoJSON",
		Run: func(cmd *cobra.Command, args []string) {
			RunConvert(circuitModel, busCoords, outputFile)
		},
	}
	convertCmd.Flags().StringVar(&circuitModel, "circuit_model", "", "Path to the circuit model file (required)")
	convertCmd.Flags().StringVar(&busCoords, "bus_coords", "", "Path to the bus coordinates file (required)")
	convertCmd.Flags().StringVar(&outputFile, "output", "output.json", "Path to the output GeoJSON file")
	convertCmd.MarkFlagRequired("circuit_model")
	convertCmd.MarkFlagRequired("bus_coords")

	var inputFile, sourceNode, targetNode string
	distanceCmd := &cobra.Command{
		Use:   "distance",
		Short: "Compute shortest path between two buses",
		Run: func(cmd *cobra.Command, args []string) {
			RunDistance(inputFile, sourceNode, targetNode)
		},
	}
	distanceCmd.Flags().StringVar(&inputFile, "input", "output.json", "Path to the GeoJSON file to read")
	distanceCmd.Flags().StringVar(&sourceNode, "source_node", "", "Source bus ID (required)")
	distanceCmd.Flags().StringVar(&targetNode, "target_node", "", "Target bus ID (required)")
	distanceCmd.MarkFlagRequired("source_node")
	distanceCmd.MarkFlagRequired("target_node")

	root.AddCommand(convertCmd, distanceCmd)
	root.SetArgs(normalizeDashes(os.Args[1:]))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
