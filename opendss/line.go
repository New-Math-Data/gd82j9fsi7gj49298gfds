package opendss

import (
	"encoding/json"

	"opendss-assessment/geojson"
)

type Line struct {
	ID    string
	Specs Specs
}

func (l *Line) ToFeature(busIndex map[BusID]*Bus) geojson.Feature {
	bus1ID := NewBusID(l.Specs.String("bus1"))
	bus2ID := NewBusID(l.Specs.String("bus2"))

	var coords json.RawMessage
	b1, ok1 := busIndex[bus1ID]
	b2, ok2 := busIndex[bus2ID]
	if ok1 && ok2 {
		coords, _ = json.Marshal([][]float64{{b1.Lon, b1.Lat}, {b2.Lon, b2.Lat}})
	}

	id := "Line." + l.ID
	return geojson.Feature{
		Type:     "Feature",
		Geometry: geojson.Geometry{Type: "LineString", Coordinates: coords},
		Properties: geojson.Properties{
			ID:              id,
			Name:            id,
			AssetType:       "Line",
			Specifications:  l.Specs,
			GlossaryTerms:   []string{},
			ConnectedAssets: geojson.ConnectedAssets{Sources: []string{string(bus1ID)}, Targets: []string{string(bus2ID)}},
		},
	}
}
