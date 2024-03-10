package logmem

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeLog_zero_division(t *testing.T) {
	t.Parallel()

	nodeLog := New(12345)

	require.InDelta(t, float64(0), nodeLog.Predict(1, 2), 0)
	require.InDelta(t, float64(0), nodeLog.PriorPfromAtoB(1, 2), 0)
	require.InDelta(t, float64(0), nodeLog.PriorPNotFromAtoB(1, 2), 0)
	require.InDelta(t, float64(0), nodeLog.PriorPtoB(1), 0)
	require.NotPanics(t, func() { nodeLog.Update(1, 2) })
}
