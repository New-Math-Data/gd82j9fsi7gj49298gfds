package opendss

import "opendss-assessment/geojson"

type Line struct {
	ID    string
	Bus1  BusID
	Bus2  BusID
	Specs Specs
}

// Convert Line from OpenDSS to GeoJSON.
func (l *Line) ToFeature(busIndex map[BusID]*Bus) *geojson.Feature {
	b1, ok1 := busIndex[l.Bus1]
	b2, ok2 := busIndex[l.Bus2]

	var geom geojson.Geometry
	if ok1 && ok2 {
		geom = geojson.Geometry(*geojson.NewLineString(geojson.NewPoint(b1.Lon, b1.Lat), geojson.NewPoint(b2.Lon, b2.Lat)))
	}

	id := "Line." + l.ID
	return geojson.NewFeature(&geom, &geojson.Properties{
		ID:              id,
		Name:            id,
		AssetType:       "Line",
		Specifications:  l.Specs,
		GlossaryTerms:   []string{},
		ConnectedAssets: geojson.ConnectedAssets{Sources: []string{string(l.Bus1)}, Targets: []string{string(l.Bus2)}},
	})
}
