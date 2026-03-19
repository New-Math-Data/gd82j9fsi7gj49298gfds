package geojson

import "encoding/json"

type Point Geometry

func NewPoint(lon, lat float64) Point {
	coords, _ := json.Marshal([2]float64{lon, lat})
	return Point{Type: "Point", Coordinates: coords}
}

type LineString Geometry

func NewLineString(points ...Point) LineString {
	raw := make([]json.RawMessage, len(points))
	for i, p := range points {
		raw[i] = p.Coordinates
	}
	coords, _ := json.Marshal(raw)
	return LineString{Type: "LineString", Coordinates: coords}
}
