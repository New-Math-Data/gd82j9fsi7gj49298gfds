package opendss

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const rawVsourceOpenDSS = `Edit "Vsource.source" bus1=tyn201 basekv=12.47 pu=1.00000001989 angle=0.000000 Z1=[0.141506, 1.399508] Z0=[0.054425, 1.088506]`

func TestNewVsourceFromOpenDSS(t *testing.T) {
	v := NewVsourceFromOpenDSS(rawVsourceOpenDSS)
	require.NotNil(t, v)

	assert.Equal(t, "source", v.ID)
	assert.Equal(t, BusID("tyn201"), v.BusID)

	assert.Equal(t, "tyn201", v.Specs.String("bus1"))
	assert.Equal(t, "12.47", v.Specs.String("basekv"))
	assert.Equal(t, "1.00000001989", v.Specs.String("pu"))
	assert.Equal(t, "0.000000", v.Specs.String("angle"))
	assert.Equal(t, "[0.141506, 1.399508]", v.Specs.String("Z1"))
	assert.Equal(t, "[0.054425, 1.088506]", v.Specs.String("Z0"))
}

func TestNewVsourceRejectsNonVsource(t *testing.T) {
	assert.Nil(t, NewVsourceFromOpenDSS(`New "Line.foo" bus1=x`))
	assert.Nil(t, NewVsourceFromOpenDSS(""))
}

func TestOpenDSSVsourceToGeoJSONVsource(t *testing.T) {
	v := NewVsourceFromOpenDSS(rawVsourceOpenDSS)
	require.NotNil(t, v)

	busMap := map[BusID]*Bus{
		BusID("tyn201"): {ID: BusID("tyn201"), Lat: 35.05225, Lon: -85.1481},
		BusID("other"):  {ID: BusID("other"), Lat: 0, Lon: 0}, // Unrelated - ensure we ignore
	}

	f := v.ToGeoJSONFeature(busMap)
	require.NotNil(t, f)

	assert.Equal(t, "Feature", f.Type)
	assert.Equal(t, "Point", f.Geometry.Type)
	assert.JSONEq(t, `[-85.1481, 35.05225]`, string(f.Geometry.Coordinates))

	assert.Equal(t, "Vsource.source", f.Properties.ID)
	assert.Equal(t, "Vsource.source", f.Properties.Name)
	assert.Equal(t, "Vsource", f.Properties.AssetType)

	assert.Equal(t, []string{}, f.Properties.ConnectedAssets.Sources)
	assert.Equal(t, []string{"tyn201"}, f.Properties.ConnectedAssets.Targets)

	assert.Equal(t, []string{"POWERFLOW"}, f.Properties.GlossaryTerms)

	assert.Equal(t, "tyn201", f.Properties.Specifications["bus1"])
	assert.Equal(t, "12.47", f.Properties.Specifications["basekv"])
	assert.Equal(t, "1.00000001989", f.Properties.Specifications["pu"])
	assert.Equal(t, "0.000000", f.Properties.Specifications["angle"])
	assert.Equal(t, "[0.141506, 1.399508]", f.Properties.Specifications["Z1"])
	assert.Equal(t, "[0.054425, 1.088506]", f.Properties.Specifications["Z0"])
}
