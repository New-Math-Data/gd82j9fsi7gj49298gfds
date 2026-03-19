package opendss

import (
	"strings"

	"opendss-assessment/geojson"
)

type BusID string

// Normalizes a given Bus ID. Sometimes a Bus ID is presented in the form
// of the canonical ID followed by metadata, delimited with a period or underscore.
func NewBusID(raw string) BusID {
	// Take everything before the first period or underscore.
	if i := strings.IndexAny(raw, "._"); i >= 0 {
		return BusID(raw[:i])
	}
	return BusID(raw)
}

type Bus struct {
	ID  BusID
	Lat float64
	Lon float64
}

// Convert Bus from OpenDSS to GeoJSON.
func (b *Bus) ToFeature() *geojson.Feature {
	geom := geojson.Geometry(*geojson.NewPoint(b.Lon, b.Lat))
	return geojson.NewFeature(&geom, &geojson.Properties{
		ID:              string(b.ID),
		Name:            string(b.ID),
		AssetType:       "Bus",
		GlossaryTerms:   []string{"POWERFLOW"},
		ConnectedAssets: geojson.ConnectedAssets{Sources: []string{}, Targets: []string{}},
	})
}
