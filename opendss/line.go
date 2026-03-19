package opendss

import "opendss-assessment/geojson"

type Line struct {
	ID     string
	BusID1 BusID
	BusID2 BusID
	Specs  Specs
}

// Converts a single line from OpenDSS file into Line struct
func NewLineFromOpenDSS(rawLine string) *Line {
	id, specTokens, ok := parseOpenDSSElement(rawLine, "Line")
	if !ok {
		return nil
	}
	specs := parseSpecs(specTokens)
	return &Line{
		ID:     id,
		BusID1: NewBusID(specs.String("bus1")),
		BusID2: NewBusID(specs.String("bus2")),
		Specs:  specs,
	}
}

// Convert Line from OpenDSS to GeoJSON.
func (l *Line) ToFeature(busIndex map[BusID]*Bus) *geojson.Feature {
	b1, ok1 := busIndex[l.BusID1]
	b2, ok2 := busIndex[l.BusID2]

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
		ConnectedAssets: geojson.ConnectedAssets{Sources: []string{string(l.BusID1)}, Targets: []string{string(l.BusID2)}},
	})
}
