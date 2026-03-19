package opendss

import (
	"encoding/json"

	"opendss-assessment/geojson"
)

type Vsource struct {
	ID    string
	Specs Specs
}

func (v *Vsource) ToFeature(busIndex map[BusID]*Bus) geojson.Feature {
	bus1ID := NewBusID(v.Specs.String("bus1"))

	var coords json.RawMessage
	if b, ok := busIndex[bus1ID]; ok {
		coords, _ = json.Marshal([]float64{b.Lon, b.Lat})
	}

	id := "Vsource." + v.ID
	return geojson.Feature{
		Type:     "Feature",
		Geometry: geojson.Geometry{Type: "Point", Coordinates: coords},
		Properties: geojson.Properties{
			ID:              id,
			Name:            id,
			AssetType:       "Vsource",
			Specifications:  v.Specs,
			GlossaryTerms:   []string{"POWERFLOW"},
			ConnectedAssets: geojson.ConnectedAssets{Sources: []string{}, Targets: []string{string(bus1ID)}},
		},
	}
}
