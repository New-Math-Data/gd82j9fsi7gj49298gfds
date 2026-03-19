#!/usr/bin/env bash
set -euo pipefail

BIN=bin/opendss-assessment
GEOJSON=results/output.json
DISTANCES=results/distance.txt

echo "==> convert: OpenDSS -> GeoJSON"
$BIN convert \
  --circuit_model data/master.dss \
  --bus_coords data/busGISCoords.csv \
  --output "$GEOJSON"

echo ""
echo "==> distance queries (using $GEOJSON) -> $DISTANCES"

run_distance() {
  local source=$1 target=$2
  $BIN distance --input "$GEOJSON" --source_node "$source" --target_node "$target" 2>&1 || true
}

run_error() {
  local desc=$1; shift
  echo "ERROR CASE: $desc"
  "$@" 2>&1 || true
}

{
  echo "--- happy path ---"
  echo ""

  echo "# expected: distance 0 (same node)"
  run_distance tyn201 tyn201
  echo ""

  echo "# expected: distance 1 (adjacent via Line.tyn201_feeder)"
  run_distance tyn201 1147583
  echo ""

  echo "# expected: distance 2 (README example)"
  run_distance tyn201 64925
  echo ""

  echo "# expected: distance 1 (same edge, opposite direction)"
  run_distance 1147583 64925
  echo ""

  echo "# expected: distance 12 (longer path)"
  run_distance tyn201 64910
  echo ""

  echo "--- error cases ---"
  echo ""

  echo "# expected: error — source node does not exist in graph"
  run_error "unknown source node" \
    $BIN distance --input "$GEOJSON" --source_node DOES_NOT_EXIST --target_node tyn201
  echo ""

  echo "# expected: error — target node does not exist in graph"
  run_error "unknown target node" \
    $BIN distance --input "$GEOJSON" --source_node tyn201 --target_node DOES_NOT_EXIST
  echo ""

  echo "# expected: error — input GeoJSON file not found"
  run_error "missing input file" \
    $BIN distance --input no_such_file.json --source_node tyn201 --target_node 64925
  echo ""

  echo "# expected: error — required flag source_node missing"
  run_error "missing --source_node flag" \
    $BIN distance --input "$GEOJSON" --target_node 64925
  echo ""

  echo "# expected: error — circuit_model flag missing from convert"
  run_error "missing --circuit_model flag" \
    $BIN convert --bus_coords data/busGISCoords.csv

} | tee "$DISTANCES"
