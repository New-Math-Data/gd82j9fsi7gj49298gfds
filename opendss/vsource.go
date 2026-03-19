package opendss

import "opendss-assessment/geojson"

type Vsource struct {
	ID    string
	BusID BusID
	Specs Specs
}

// Converts a single Vsource from OpenDSS file into Vsource struct
func NewVsourceFromOpenDSS(rawLine string) *Vsource {
	id, specTokens, ok := parseOpenDSSElement(rawLine, "Vsource")
	if !ok {
		return nil
	}
	specs := parseSpecs(specTokens)
	return &Vsource{
		ID:    id,
		BusID: NewBusID(specs.String("bus1")),
		Specs: specs,
	}
}

// Convert Vsource from OpenDSS to GeoJSON.
func (v *Vsource) ToFeature(busIndex map[BusID]*Bus) *geojson.Feature {
	var geom geojson.Geometry
	if b, ok := busIndex[v.BusID]; ok {
		geom = geojson.Geometry(*geojson.NewPoint(b.Lon, b.Lat))
	}

	id := "Vsource." + v.ID
	return geojson.NewFeature(&geom, &geojson.Properties{
		ID:              id,
		Name:            id,
		AssetType:       "Vsource",
		Specifications:  v.Specs,
		GlossaryTerms:   []string{"POWERFLOW"},
		ConnectedAssets: geojson.ConnectedAssets{Sources: []string{}, Targets: []string{string(v.BusID)}},
	})
}
