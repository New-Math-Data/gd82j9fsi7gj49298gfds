package opendss

import "opendss-assessment/geojson"

type Line struct {
	ID    string
	Specs Specs
}

func (l *Line) ToFeature(busIndex map[BusID]*Bus) *geojson.Feature {
	bus1ID := NewBusID(l.Specs.String("bus1"))
	bus2ID := NewBusID(l.Specs.String("bus2"))

	b1, ok1 := busIndex[bus1ID]
	b2, ok2 := busIndex[bus2ID]

	var geom geojson.Geometry
	if ok1 && ok2 {
		geom = geojson.Geometry(geojson.NewLineString(geojson.NewPoint(b1.Lon, b1.Lat), geojson.NewPoint(b2.Lon, b2.Lat)))
	}

	id := "Line." + l.ID
	return &geojson.Feature{
		Type:     "Feature",
		Geometry: geom,
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
