package opendss

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVsource(t *testing.T) {
	raw := `Edit "Vsource.source" bus1=tyn201 basekv=12.47 pu=1.00000001989 angle=0.000000 Z1=[0.141506, 1.399508] Z0=[0.054425, 1.088506]`
	v := NewVsourceFromOpenDSS(raw)
	require.NotNil(t, v)
	assert.Equal(t, "source", v.ID)
	assert.Equal(t, BusID("tyn201"), v.BusID)
	assert.Equal(t, "12.47", v.Specs.String("basekv"))
}

func TestNewVsourceRejectsNonVsource(t *testing.T) {
	assert.Nil(t, NewVsourceFromOpenDSS(`New "Line.foo" bus1=x`))
	assert.Nil(t, NewVsourceFromOpenDSS(""))
}
