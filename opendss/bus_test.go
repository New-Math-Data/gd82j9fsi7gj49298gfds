package opendss

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBusFromCSV(t *testing.T) {
	b := NewBusFromCSV([]string{"tyn201", "35.05225", "-85.1481"})
	require.NotNil(t, b)
	assert.Equal(t, BusID("tyn201"), b.ID)
	assert.Equal(t, 35.05225, b.Lat)
	assert.Equal(t, -85.1481, b.Lon)
}

func TestNewBusShortRow(t *testing.T) {
	assert.Nil(t, NewBusFromCSV([]string{"tyn201", "35.05225"}))
}

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

func TestLoadBusCoords(t *testing.T) {
	c := NewCircuit()
	require.NoError(t, c.LoadBusCoords("../data/busGISCoords.csv"))
	assert.NotEmpty(t, c.buses)

	b, ok := c.buses[BusID("tyn201")]
	require.True(t, ok, "tyn201 not found in busIndex")
	assert.Equal(t, 35.05225, b.Lat)
	assert.Equal(t, -85.1481, b.Lon)
}

func TestOpenDSSBusToGeoJSONBus(t *testing.T) {
	b := NewBusFromCSV([]string{"tyn201", "35.05225", "-85.1481"})
	require.NotNil(t, b)

	f := b.ToGeoJsonFeature()
	require.NotNil(t, f)

	assert.Equal(t, "Feature", f.Type)
	assert.Equal(t, "Point", f.Geometry.Type)
	assert.JSONEq(t, `[-85.1481, 35.05225]`, string(f.Geometry.Coordinates))

	assert.Equal(t, "tyn201", f.Properties.ID)
	assert.Equal(t, "tyn201", f.Properties.Name)
	assert.Equal(t, "Bus", f.Properties.AssetType)

	assert.Equal(t, []string{}, f.Properties.ConnectedAssets.Sources)
	assert.Equal(t, []string{}, f.Properties.ConnectedAssets.Targets)

	assert.Equal(t, []string{"POWERFLOW"}, f.Properties.GlossaryTerms)
}
