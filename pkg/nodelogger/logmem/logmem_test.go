package logmem

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeLog_zero_division(t *testing.T) {
	t.Parallel()

	nodeLog := New(12345)

	require.Equal(t, nodeLog.Predict(1, 2), float64(0))
	require.Equal(t, nodeLog.PriorPfromAtoB(1, 2), float64(0))
	require.Equal(t, nodeLog.PriorPNotFromAtoB(1, 2), float64(0))
	require.Equal(t, nodeLog.PriorPtoB(1), float64(0))
	require.NotPanics(t, func() { nodeLog.Update(1, 2) })
}
