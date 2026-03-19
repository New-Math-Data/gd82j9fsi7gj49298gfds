# Backend Engineering Take-Home Assessment

## Introduction

This is a take-home assessment designed to evaluate your backend engineering and graph data skills using Go. You have **24 hours** to submit your solution in the form of a PR.

You can use any available tools at your disposal, including documentation, LLMs, or AI assistants as you would on a normal day-to-day basis. However, this is a solo exercise.

We don't expect you to be an expert in all of the concepts used here. Part of the exercise is to show your ability to look up and apply common patterns in unfamiliar domains.

The problem is judged on the following criteria, in order of importance:

1. Submission of running Go code
2. Correct GeoJSON output (valid FeatureCollection with proper geometry types, coordinates, and connectivity)
3. Correct shortest path computation
4. Code quality, including readability, efficiency, and use of Go conventions

Submitting a solution that meets some of the requirements is preferred to incomplete code.

---

## Background

**OpenDSS** is a widely-used open-source tool for simulating electric power distribution systems. It uses plain-text files to describe circuit models (power lines, voltage sources, buses, etc.) and coordinate data for those components.

**GeoJSON** is a standard format for encoding geographic data structures. Our API ingests GeoJSON `FeatureCollection` documents to model infrastructure networks.

Your task is to bridge these two worlds: parse OpenDSS data and produce GeoJSON that our system can consume, then use that data to answer graph queries.

---

## Problem

You are given two data files in the `data/` directory:

- **`busGISCoords.csv`** -- Bus (node) coordinates in `bus_id,latitude,longitude` format
- **`master.dss`** -- An OpenDSS circuit model containing voltage sources, wire data, line spacing, and power line definitions

### Part 1: OpenDSS to GeoJSON Conversion

Write a Go program that:

1. Parses `busGISCoords.csv` into GeoJSON **Point** features representing **Buses**
2. Parses `master.dss` to extract **Line** definitions (power lines connecting two buses) and **Vsource** definitions (voltage sources connected to a bus)
3. Builds connectivity between these assets via the `connected_assets` field
4. Resolves geometries:
   - **Bus**: Point geometry from the coordinates file
   - **Line**: LineString geometry derived from the coordinates of its two endpoint buses
   - **Vsource**: Point geometry copied from its connected bus
5. Outputs a valid GeoJSON `FeatureCollection` to `output.json`

### Part 2: Graph Construction & Shortest Path

Implement a second command that:

1. Reads the GeoJSON `FeatureCollection` produced by the `convert` command
2. Builds an **in-memory graph** where **Buses are nodes** and **Lines are edges** (undirected, unweighted)
3. Accepts two bus IDs via command-line flags (`-source_node` and `-target_node`)
4. Computes and prints the **shortest distance** (hop count) between the two buses
5. Prints `-1` if no path exists

---

## GeoJSON Schema

Your output must conform to this structure:

```json
{
  "type": "FeatureCollection",
  "features": [...]
}
```

Each feature must have this shape:

```json
{
  "type": "Feature",
  "geometry": {
    "type": "<Point|LineString>",
    "coordinates": "<[lon, lat] or [[lon, lat], ...]>"
  },
  "properties": {
    "id": "<string>",
    "name": "<string>",
    "assetType": "<Bus|Line|Vsource>",
    "specifications": {},
    "connected_assets": {
      "sources": [],
      "targets": []
    },
    "glossary_terms": []
  }
}
```

**Important**: GeoJSON coordinates are `[longitude, latitude]`, not `[latitude, longitude]`.

---

## Expected Output Examples

Given the input data provided, here is what each asset type should look like when converted:

### Bus (Point)

Input line from `busGISCoords.csv`:
```
tyn201,35.05225,-85.1481
```

Expected GeoJSON feature:
```json
{
  "type": "Feature",
  "geometry": {
    "type": "Point",
    "coordinates": [-85.1481, 35.05225]
  },
  "properties": {
    "id": "tyn201",
    "name": "tyn201",
    "assetType": "Bus",
    "connected_assets": {
      "sources": ["Vsource.source"],
      "targets": ["Line.tyn201_feeder"]
    },
    "glossary_terms": ["POWERFLOW"]
  }
}
```

Note: The `connected_assets` above reflect the fully wired state. Bus `tyn201` has the voltage source `Vsource.source` as an incoming source, and `Line.tyn201_feeder` as an outgoing target (since the feeder line's `bus1` is `tyn201`).

### Line (LineString)

Input line from `master.dss`:
```
New "Line.tyn201_feeder" phases=3 bus1=tyn201.1.2.3 bus2=1147583.1.2.3 length=30.7848 units=m spacing=3PH_HORIZ_LG_1C_3 wires="AL_556_19STR AL_556_19STR AL_556_19STR ACSR_1/0_6/1" Seasons=1 Ratings=[400,] normamps=613 emergamps=613
```

Expected GeoJSON feature:
```json
{
  "type": "Feature",
  "geometry": {
    "type": "LineString",
    "coordinates": [
      [-85.1481, 35.05225],
      [-85.14808268, 35.05217897]
    ]
  },
  "properties": {
    "id": "Line.tyn201_feeder",
    "name": "Line.tyn201_feeder",
    "assetType": "Line",
    "specifications": {
      "phases": "3",
      "bus1": "tyn201.1.2.3",
      "bus2": "1147583.1.2.3",
      "length": "30.7848",
      "units": "m",
      "spacing": "3PH_HORIZ_LG_1C_3",
      "wires": "\"AL_556_19STR AL_556_19STR AL_556_19STR ACSR_1/0_6/1\"",
      "Seasons": "1",
      "Ratings": "[400,]",
      "normamps": "613",
      "emergamps": "613"
    },
    "connected_assets": {
      "sources": ["tyn201"],
      "targets": ["1147583"]
    },
    "glossary_terms": []
  }
}
```

Key points:
- The `id` is the full identifier from the DSS file (e.g. `Line.tyn201_feeder`)
- Bus references like `tyn201.1.2.3` must be normalized to just `tyn201` for `connected_assets` (split on `.` and take the first token)
- References like `108687_1s.1.2.3` should also be normalized -- split on both `_` and `.`, take the first token (yielding `108687`)
- `connected_assets.sources` contains the `bus1` ID (normalized), `connected_assets.targets` contains the `bus2` ID (normalized)
- LineString coordinates come from the source bus coordinates followed by the target bus coordinates

### Vsource (Point)

Input line from `master.dss`:
```
Edit "Vsource.source" bus1=tyn201 basekv=12.47 pu=1.00000001989 angle=0.000000 Z1=[0.141506, 1.399508] Z0=[0.054425, 1.088506]
```

Expected GeoJSON feature:
```json
{
  "type": "Feature",
  "geometry": {
    "type": "Point",
    "coordinates": [-85.1481, 35.05225]
  },
  "properties": {
    "id": "Vsource.source",
    "name": "Vsource.source",
    "assetType": "Vsource",
    "specifications": {
      "bus1": "tyn201",
      "basekv": "12.47",
      "pu": "1.00000001989",
      "angle": "0.000000",
      "Z1": "[0.141506,",
      "1.399508]": "true",
      "Z0": "[0.054425,",
      "1.088506]": "true"
    },
    "connected_assets": {
      "sources": [],
      "targets": ["tyn201"]
    },
    "glossary_terms": ["POWERFLOW"]
  }
}
```

Key points:
- Vsource geometry is copied from its connected bus (`bus1`)
- `connected_assets.targets` contains the bus the Vsource feeds into
- Vsources can start with `New` or `Edit`

---

## Parsing Hints

### Bus Coordinates (`busGISCoords.csv`)

Each line is: `bus_id,latitude,longitude`

Coordinates must be flipped for GeoJSON: output `[longitude, latitude]`.

### Circuit Model (`master.dss`)

The file contains several types of definitions. You only need to parse:

- **Lines**: rows starting with `New "Line.<id>"` -- these define power lines connecting two buses via `bus1=` and `bus2=` parameters
- **Vsources**: rows where the second field contains `"Vsource.<id>"` -- these can start with either `New` or `Edit`

You can safely ignore `WireData`, `LineSpacing`, `LineCode`, `Circuit`, `Clear`, `Buscoords`, and any other definition types.

### Bus ID Normalization

Bus references in the DSS file include phase information that must be stripped:
- `tyn201.1.2.3` -> `tyn201`
- `1001176.3` -> `1001176`
- `108687_1s.1.2.3` -> `108687`

Split on both `.` and `_`, then take the **first token** as the bus ID.

### Connectivity Wiring

After parsing all assets, wire them together:
- Each **Line** has `connected_assets.sources = [bus1_id]` and `connected_assets.targets = [bus2_id]`
- Each **Bus** accumulates references to Lines and Vsources that connect to it:
  - `sources`: assets where this bus is the **target** (incoming connections)
  - `targets`: assets where this bus is the **source** (outgoing connections)
- Each **Vsource** has `connected_assets.targets = [bus1_id]`

---

## Command-Line Interface

Your program should expose two subcommands:

### `convert` -- Convert OpenDSS to GeoJSON

```
go run main.go convert [flags]

Flags:
  -circuit_model  Path to the circuit model file (required)
  -bus_coords     Path to the bus coordinates file (required)
  -output         Path to the output GeoJSON file (default: output.json)
```

Example:
```bash
go run main.go convert -circuit_model data/master.dss -bus_coords data/busGISCoords.csv
```

This parses the OpenDSS data and writes a GeoJSON `FeatureCollection` to `output.json`.

### `distance` -- Compute Shortest Path

```
go run main.go distance [flags]

Flags:
  -input        Path to the GeoJSON file to read (default: output.json)
  -source_node  Source bus ID (required)
  -target_node  Target bus ID (required)
```

Example:
```bash
go run main.go distance -source_node tyn201 -target_node 64925
```

This reads the converted GeoJSON, builds a graph from it, and prints the shortest path distance between the two buses.

---

## Submission

Please submit:
1. All Go source files (`main.go`, `go.mod`, and any additional files)
2. The generated `output.json` file

---

## Useful Libraries

- [`encoding/json`](https://pkg.go.dev/encoding/json) -- JSON marshalling/unmarshalling (standard library)
- [`flag`](https://pkg.go.dev/flag) -- Command-line flag parsing (standard library)
- [`os`](https://pkg.go.dev/os) -- File I/O (standard library)
- [`strings`](https://pkg.go.dev/strings) -- String manipulation (standard library)
- [`regexp`](https://pkg.go.dev/regexp) -- Regular expressions (standard library)

No external dependencies are required for this assessment.

---

## Evaluation

How to navigate what I built:

- `scripts/run.sh` provides numerous runs of the app showing its output.
- `results/output.json` contains OpenDSS converted to GeoJSON.
- `results/distance.txt` contains the output of `scripts/run.sh` showing distancs calcs and some expected error cases.
- `results/claude.transcript` contains a lo-fi transcript of my conversation with Claude, which I used to write most of this.
- `make run` builds an executable and runs `scripts/run.sh`

**Note**: Did I go overboard? Maybe, but AI tools like Claude make it easy and fast to churn out test cases and productionalize (linting, Makefile, CI, etc.) code.
As an engineer, it's important that you know how to guide it to create a good architecture, ensure the core test cases are truly covered and correct, and to
check the core parts of the code (like the BFS) to ensure they accomplish what you want and they don't do things like infinite loop or memory exhaustion.
