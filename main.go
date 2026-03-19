package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"opendss-assessment/geojson"
	"opendss-assessment/graph"
	"opendss-assessment/opendss"
)

// this just allows you to use long-form command line args with one or two dashes:
// opendss-assessment distance --source_node whatever -target_node whatver
func normalizeDashes(args []string) []string {
	normalized := make([]string, len(args))
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			// you can specify with or without an equals:
			// --my_arg=35 or -my_arg 35
			name := strings.SplitN(arg[1:], "=", 2)[0]
			// covers the case of -a=35  len(name)==1
			if len(name) > 1 {
				arg = "-" + arg
			}
		}
		normalized[i] = arg
	}
	return normalized
}

func RunConvert(circuitModel, busCoords, outputFile string) error {
	circuit := opendss.NewCircuit()
	if err := circuit.LoadBusCoords(busCoords); err != nil {
		return err
	}
	if err := circuit.LoadCircuitModel(circuitModel); err != nil {
		return err
	}
	collection := circuit.ToGeoJSON()

	output, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputFile, output, 0644); err != nil {
		return err
	}

	fmt.Printf("Wrote to %s\n", outputFile)
	return nil
}

func RunDistance(inputFile, sourceNodeID, targetNodeID string) error {
	collection, err := geojson.LoadGeoJSON(inputFile)
	if err != nil {
		return err
	}
	g := collection.BuildGraph()

	sourceNode := g.Node(graph.NodeID(sourceNodeID))
	if sourceNode == nil {
		return fmt.Errorf("source node %s not found", sourceNodeID)
	}

	targetNode := g.Node(graph.NodeID(targetNodeID))
	if targetNode == nil {
		return fmt.Errorf("target node %s not found", targetNodeID)
	}

	distance := g.ShortestPath(sourceNode, targetNode)
	fmt.Printf("Shortest path from %s to %s: %d\n", sourceNodeID, targetNodeID, distance)
	return nil
}

func main() {
	// I use the cobra library for parsing the command line because it provides so
	// many built-in utilities like a help function and parsing.
	root := &cobra.Command{
		Use:   "opendss-assessment",
		Short: "OpenDSS to GeoJSON converter and graph path finder",
	}

	var circuitModel, busCoords, outputFile string
	convertCmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert OpenDSS data to GeoJSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunConvert(circuitModel, busCoords, outputFile)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunDistance(inputFile, sourceNode, targetNode)
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
		fmt.Println(err)
		os.Exit(1)
	}
}
