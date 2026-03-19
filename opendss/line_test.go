package opendss

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const rawLineOpenDSS = `New "Line.tyn201_feeder" phases=3 bus1=tyn201.1.2.3 bus2=1147583.1.2.3 length=30.7848 units=m spacing=3PH_HORIZ_LG_1C_3 wires="AL_556_19STR AL_556_19STR AL_556_19STR ACSR_1/0_6/1" Seasons=1 Ratings=[400,] normamps=613 emergamps=613`

func TestNewLineFromOpenDSS(t *testing.T) {
	l := NewLineFromOpenDSS(rawLineOpenDSS)
	require.NotNil(t, l)

	assert.Equal(t, "tyn201_feeder", l.ID)
	assert.Equal(t, BusID("tyn201"), l.BusID1)  // canonical bus ID
	assert.Equal(t, BusID("1147583"), l.BusID2) // canonical bus ID

	assert.Equal(t, "3", l.Specs.String("phases"))
	assert.Equal(t, "tyn201.1.2.3", l.Specs.String("bus1"))  // bus ID with metadata
	assert.Equal(t, "1147583.1.2.3", l.Specs.String("bus2")) // bus ID with metadata
	assert.Equal(t, "30.7848", l.Specs.String("length"))
	assert.Equal(t, "m", l.Specs.String("units"))
	assert.Equal(t, "3PH_HORIZ_LG_1C_3", l.Specs.String("spacing"))
	assert.Equal(t, `"AL_556_19STR AL_556_19STR AL_556_19STR ACSR_1/0_6/1"`, l.Specs.String("wires"))
	assert.Equal(t, "1", l.Specs.String("Seasons"))
	assert.Equal(t, "[400,]", l.Specs.String("Ratings"))
	assert.Equal(t, "613", l.Specs.String("normamps"))
	assert.Equal(t, "613", l.Specs.String("emergamps"))
}

func TestNewLineFromOpenDSSRejectsNonLine(t *testing.T) {
	assert.Nil(t, NewLineFromOpenDSS(`New "Transformer.foo" phases=3`))
	assert.Nil(t, NewLineFromOpenDSS(""))
}

func TestOpenDSSLineToGeoJSONLine(t *testing.T) {
	openDSSLine := NewLineFromOpenDSS(rawLineOpenDSS)
	require.NotNil(t, openDSSLine)

	busMap := map[BusID]*Bus{
		BusID("tyn201"):  {ID: BusID("tyn201"), Lat: 35.05225, Lon: -85.1481},
		BusID("1147583"): {ID: BusID("1147583"), Lat: 35.05217897, Lon: -85.14808268},
		BusID("74056"):   {ID: BusID("74056"), Lat: 0, Lon: 0}, // Unrelated - ensure we ignore
	}

	f := openDSSLine.ToGeoJSONFeature(busMap)
	require.NotNil(t, f)

	assert.Equal(t, "Feature", f.Type)
	assert.Equal(t, "LineString", f.Geometry.Type)
	assert.JSONEq(t, `[[-85.1481,35.05225],[-85.14808268,35.05217897]]`, string(f.Geometry.Coordinates))

	assert.Equal(t, "Line.tyn201_feeder", f.Properties.ID)
	assert.Equal(t, "Line.tyn201_feeder", f.Properties.Name)
	assert.Equal(t, "Line", f.Properties.AssetType)

	assert.Equal(t, []string{"tyn201"}, f.Properties.ConnectedAssets.Sources)  // canonical bus ID
	assert.Equal(t, []string{"1147583"}, f.Properties.ConnectedAssets.Targets) // canonical bus ID

	assert.Equal(t, []string{}, f.Properties.GlossaryTerms)

	assert.Equal(t, "3", f.Properties.Specifications["phases"])
	assert.Equal(t, "tyn201.1.2.3", f.Properties.Specifications["bus1"])  // bus ID with metadata
	assert.Equal(t, "1147583.1.2.3", f.Properties.Specifications["bus2"]) // bus ID with metadata

	assert.Equal(t, "30.7848", f.Properties.Specifications["length"])
	assert.Equal(t, "m", f.Properties.Specifications["units"])
	assert.Equal(t, "3PH_HORIZ_LG_1C_3", f.Properties.Specifications["spacing"])
	assert.Equal(t, `"AL_556_19STR AL_556_19STR AL_556_19STR ACSR_1/0_6/1"`, f.Properties.Specifications["wires"])
	assert.Equal(t, "1", f.Properties.Specifications["Seasons"])
	assert.Equal(t, "[400,]", f.Properties.Specifications["Ratings"])
	assert.Equal(t, "613", f.Properties.Specifications["normamps"])
	assert.Equal(t, "613", f.Properties.Specifications["emergamps"])
}
