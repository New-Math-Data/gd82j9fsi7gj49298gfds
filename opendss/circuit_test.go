package opendss

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendss-assessment/geojson"
)

func TestParseSpecsQuotedMultiToken(t *testing.T) {
	tokens := []string{`wires="AL_556_19STR`, `AL_556_19STR`, `AL_556_19STR`, `ACSR_1/0_6/1"`, `normamps=613`}
	specs := parseSpecs(tokens)
	assert.Equal(t, `"AL_556_19STR AL_556_19STR AL_556_19STR ACSR_1/0_6/1"`, specs["wires"])
	assert.Equal(t, "613", specs["normamps"])
}

func TestParseSpecsBracketArrayTokens(t *testing.T) {
	tokens := []string{`Z1=[0.141506,`, `1.399508]`}
	specs := parseSpecs(tokens)
	assert.Equal(t, "[0.141506, 1.399508]", specs["Z1"])
	assert.NotContains(t, specs, "1.399508]")
}

func TestLoadCircuitModel(t *testing.T) {
	c := NewCircuit()
	require.NoError(t, c.LoadCircuitModel("../data/master.dss"))
	assert.NotEmpty(t, c.lines)
	assert.NotEmpty(t, c.vsources)

	var feeder *Line
	for _, l := range c.lines {
		if l.ID == "tyn201_feeder" {
			feeder = l
			break
		}
	}
	require.NotNil(t, feeder, "tyn201_feeder not found")

	assert.Equal(t, "tyn201_feeder", feeder.ID)
	assert.Equal(t, BusID("tyn201"), feeder.BusID1)
	assert.Equal(t, BusID("1147583"), feeder.BusID2)

	assert.Equal(t, "3", feeder.Specs.String("phases"))
	assert.Equal(t, "tyn201.1.2.3", feeder.Specs.String("bus1"))
	assert.Equal(t, "1147583.1.2.3", feeder.Specs.String("bus2"))
	assert.Equal(t, "30.7848", feeder.Specs.String("length"))
	assert.Equal(t, "m", feeder.Specs.String("units"))
	assert.Equal(t, "3PH_HORIZ_LG_1C_3", feeder.Specs.String("spacing"))
	assert.Equal(t, `"AL_556_19STR AL_556_19STR AL_556_19STR ACSR_1/0_6/1"`, feeder.Specs.String("wires"))
	assert.Equal(t, "1", feeder.Specs.String("Seasons"))
	assert.Equal(t, "[400,]", feeder.Specs.String("Ratings"))
	assert.Equal(t, "613", feeder.Specs.String("normamps"))
	assert.Equal(t, "613", feeder.Specs.String("emergamps"))

	var vsrc *Vsource
	for _, v := range c.vsources {
		if v.ID == "source" {
			vsrc = v
			break
		}
	}
	require.NotNil(t, vsrc, "Vsource.source not found")

	assert.Equal(t, "source", vsrc.ID)
	assert.Equal(t, BusID("tyn201"), vsrc.BusID)

	assert.Equal(t, "tyn201", vsrc.Specs.String("bus1"))
	assert.Equal(t, "12.47", vsrc.Specs.String("basekv"))
	assert.Equal(t, "1.00000001989", vsrc.Specs.String("pu"))
	assert.Equal(t, "0.000000", vsrc.Specs.String("angle"))
	assert.Equal(t, "[0.141506, 1.399508]", vsrc.Specs.String("Z1"))
	assert.Equal(t, "[0.054425, 1.088506]", vsrc.Specs.String("Z0"))
}

func TestToGeoJSON(t *testing.T) {
	c := NewCircuit()
	require.NoError(t, c.LoadBusCoords("../data/busGISCoords.csv"))
	require.NoError(t, c.LoadCircuitModel("../data/master.dss"))
	collection := c.ToGeoJSON()

	var feeder, tyn201Bus, vsrcFeature *geojson.Feature
	for _, f := range collection.Features {
		switch f.Properties.ID {
		case "Line.tyn201_feeder":
			feeder = f
		case "tyn201":
			tyn201Bus = f
		case "Vsource.source":
			vsrcFeature = f
		}
	}

	require.NotNil(t, feeder, "Line.tyn201_feeder not found")

	assert.Equal(t, "Feature", feeder.Type)
	assert.Equal(t, "LineString", feeder.Geometry.Type)
	assert.JSONEq(t, `[[-85.1481,35.05225],[-85.14808268,35.05217897]]`, string(feeder.Geometry.Coordinates))

	assert.Equal(t, "Line.tyn201_feeder", feeder.Properties.ID)
	assert.Equal(t, "Line.tyn201_feeder", feeder.Properties.Name)
	assert.Equal(t, "Line", feeder.Properties.AssetType)
	assert.Equal(t, []string{"tyn201"}, feeder.Properties.ConnectedAssets.Sources)
	assert.Equal(t, []string{"1147583"}, feeder.Properties.ConnectedAssets.Targets)
	assert.Equal(t, []string{}, feeder.Properties.GlossaryTerms)
	assert.Equal(t, "3", feeder.Properties.Specifications["phases"])
	assert.Equal(t, "tyn201.1.2.3", feeder.Properties.Specifications["bus1"])
	assert.Equal(t, "1147583.1.2.3", feeder.Properties.Specifications["bus2"])
	assert.Equal(t, "30.7848", feeder.Properties.Specifications["length"])
	assert.Equal(t, "m", feeder.Properties.Specifications["units"])
	assert.Equal(t, "613", feeder.Properties.Specifications["normamps"])
	assert.Equal(t, "613", feeder.Properties.Specifications["emergamps"])

	require.NotNil(t, vsrcFeature, "Vsource.source not found")

	assert.Equal(t, "Feature", vsrcFeature.Type)
	assert.Equal(t, "Point", vsrcFeature.Geometry.Type)
	assert.JSONEq(t, `[-85.1481, 35.05225]`, string(vsrcFeature.Geometry.Coordinates))

	assert.Equal(t, "Vsource.source", vsrcFeature.Properties.ID)
	assert.Equal(t, "Vsource.source", vsrcFeature.Properties.Name)
	assert.Equal(t, "Vsource", vsrcFeature.Properties.AssetType)
	assert.Equal(t, []string{}, vsrcFeature.Properties.ConnectedAssets.Sources)
	assert.Equal(t, []string{"tyn201"}, vsrcFeature.Properties.ConnectedAssets.Targets)
	assert.Equal(t, []string{"POWERFLOW"}, vsrcFeature.Properties.GlossaryTerms)
	assert.Equal(t, "tyn201", vsrcFeature.Properties.Specifications["bus1"])
	assert.Equal(t, "12.47", vsrcFeature.Properties.Specifications["basekv"])
	assert.Equal(t, "1.00000001989", vsrcFeature.Properties.Specifications["pu"])
	assert.Equal(t, "0.000000", vsrcFeature.Properties.Specifications["angle"])
	assert.Equal(t, "[0.141506, 1.399508]", vsrcFeature.Properties.Specifications["Z1"])
	assert.Equal(t, "[0.054425, 1.088506]", vsrcFeature.Properties.Specifications["Z0"])

	require.NotNil(t, tyn201Bus, "tyn201 bus not found")

	assert.Equal(t, "Feature", tyn201Bus.Type)
	assert.Equal(t, "Point", tyn201Bus.Geometry.Type)
	assert.JSONEq(t, `[-85.1481,35.05225]`, string(tyn201Bus.Geometry.Coordinates))

	assert.Equal(t, "tyn201", tyn201Bus.Properties.ID)
	assert.Equal(t, "tyn201", tyn201Bus.Properties.Name)
	assert.Equal(t, "Bus", tyn201Bus.Properties.AssetType)
	assert.Equal(t, []string{"POWERFLOW"}, tyn201Bus.Properties.GlossaryTerms)
	assert.Contains(t, tyn201Bus.Properties.ConnectedAssets.Sources, "Vsource.source")
	assert.Contains(t, tyn201Bus.Properties.ConnectedAssets.Targets, "Line.tyn201_feeder")
}
