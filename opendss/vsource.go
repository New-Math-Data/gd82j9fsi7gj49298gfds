package opendss

import "opendss-assessment/geojson"

type Vsource struct {
	ID    string
	Bus1  BusID
	Specs Specs
}

// Convert Vsource from OpenDSS to GeoJSON.
func (v *Vsource) ToFeature(busIndex map[BusID]*Bus) *geojson.Feature {
	var geom geojson.Geometry
	if b, ok := busIndex[v.Bus1]; ok {
		geom = geojson.Geometry(*geojson.NewPoint(b.Lon, b.Lat))
	}

	id := "Vsource." + v.ID
	return geojson.NewFeature(&geom, &geojson.Properties{
		ID:              id,
		Name:            id,
		AssetType:       "Vsource",
		Specifications:  v.Specs,
		GlossaryTerms:   []string{"POWERFLOW"},
		ConnectedAssets: geojson.ConnectedAssets{Sources: []string{}, Targets: []string{string(v.Bus1)}},
	})
}
