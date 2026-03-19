package opendss

import (
	"encoding/json"
	"regexp"

	"opendss-assessment/geojson"
)

type BusID string

var busIDRe = regexp.MustCompile(`^(\d+)[_.]|^([^.]+)`)

func NewBusID(raw string) BusID {
	m := busIDRe.FindStringSubmatch(raw)
	if m == nil {
		return BusID(raw)
	}
	if m[1] != "" {
		return BusID(m[1])
	}
	return BusID(m[2])
}

type Bus struct {
	ID  BusID
	Lat float64
	Lon float64
}

func (b *Bus) ToFeature() geojson.Feature {
	coords, _ := json.Marshal([]float64{b.Lon, b.Lat})
	return geojson.Feature{
		Type:     "Feature",
		Geometry: geojson.Geometry{Type: "Point", Coordinates: coords},
		Properties: geojson.Properties{
			ID:              string(b.ID),
			Name:            string(b.ID),
			AssetType:       "Bus",
			GlossaryTerms:   []string{"POWERFLOW"},
			ConnectedAssets: geojson.ConnectedAssets{Sources: []string{}, Targets: []string{}},
		},
	}
}
