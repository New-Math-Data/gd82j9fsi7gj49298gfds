package opendss

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendss-assessment/geojson"
)

func TestNewBusID(t *testing.T) {
	cases := []struct {
		input string
		want  BusID
	}{
		{"tyn201.1.2.3", "tyn201"},
		{"1001176.3", "1001176"},
		{"108687_1s.1.2.3", "108687"},
		{"tyn201", "tyn201"},
		{"1147844_1s.1.2.3", "1147844"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, NewBusID(c.input))
	}
}

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

func TestLoadBusCoords(t *testing.T) {
	c := NewCircuit()
	require.NoError(t, c.LoadBusCoords("../data/busGISCoords.csv"))
	assert.NotEmpty(t, c.busIndex)

	b, ok := c.busIndex[BusID("tyn201")]
	require.True(t, ok, "tyn201 not found in busIndex")
	assert.Equal(t, 35.05225, b.Lat)
	assert.Equal(t, -85.1481, b.Lon)
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
	assert.Equal(t, "tyn201.1.2.3", feeder.Specs.String("bus1"))

	var vsrc *Vsource
	for _, v := range c.vsources {
		if v.ID == "source" {
			vsrc = v
			break
		}
	}
	require.NotNil(t, vsrc, "Vsource.source not found")
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
	assert.Equal(t, []string{"tyn201"}, feeder.Properties.ConnectedAssets.Sources)
	assert.Equal(t, []string{"1147583"}, feeder.Properties.ConnectedAssets.Targets)

	require.NotNil(t, vsrcFeature, "Vsource.source not found")
	assert.Equal(t, []string{"tyn201"}, vsrcFeature.Properties.ConnectedAssets.Targets)

	require.NotNil(t, tyn201Bus, "tyn201 bus not found")
	assert.Contains(t, tyn201Bus.Properties.ConnectedAssets.Sources, "Vsource.source")
}
