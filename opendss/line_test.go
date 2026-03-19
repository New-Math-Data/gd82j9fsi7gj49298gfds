package opendss

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLine(t *testing.T) {
	raw := `New "Line.bon203_74056usd" phases=3 bus1=108687_1s.1.2.3 bus2=74056.1.2.3 length=1.524 units=m spacing=3PH_HORIZ_LG_1C_3 wires="BUSBAR BUSBAR BUSBAR BUSBAR" Seasons=1 Ratings=[400,] normamps=2000 emergamps=2000`
	l := NewLineFromOpenDSS(raw)
	require.NotNil(t, l)
	assert.Equal(t, "bon203_74056usd", l.ID)
	assert.Equal(t, BusID("108687"), l.BusID1)
	assert.Equal(t, BusID("74056"), l.BusID2)
	assert.Equal(t, "3", l.Specs.String("phases"))
}

func TestNewLineRejectsNonLine(t *testing.T) {
	assert.Nil(t, NewLineFromOpenDSS(`New "Transformer.foo" phases=3`))
	assert.Nil(t, NewLineFromOpenDSS(""))
}
