package logmem

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeLog_zero_division(t *testing.T) {
	n := New(12345)

	require.Equal(t, n.Predict(1, 2), float64(0))
	require.Equal(t, n.PriorPfromAtoB(1, 2), float64(0))
	require.Equal(t, n.PriorPNotFromAtoB(1, 2), float64(0))
	require.Equal(t, n.PriorPtoB(1), float64(0))
	require.NotPanics(t, func() { n.Update(1, 2) })
}
