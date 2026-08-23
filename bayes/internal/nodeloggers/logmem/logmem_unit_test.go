package logmem

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeLog_zero_division(t *testing.T) {
	t.Parallel()

	nodeLog := New(12345)

	require.InDelta(t, float64(0), nodeLog.Predict(1, 2), 0)
	require.InDelta(t, float64(0), nodeLog.PriorProbFromTo(1, 2), 0)
	require.InDelta(t, float64(0), nodeLog.PriorProbNotFromTo(1, 2), 0)
	require.InDelta(t, float64(0), nodeLog.PriorProbTo(1), 0)
	require.NotPanics(t, func() { nodeLog.Update(1, 2) })
}

func TestSnapshot_deepCopyIsolation(t *testing.T) {
	t.Parallel()

	want := Snapshot{
		NodeID:        42,
		TotalAccesses: 3,
		FromA:         map[uint64]int{1: 2},
		ToB:           map[uint64]int{2: 2},
		FromAToB: map[uint64]map[uint64]int{
			1: {2: 2},
		},
	}
	input := Snapshot{
		NodeID:        42,
		TotalAccesses: 3,
		FromA:         map[uint64]int{1: 2},
		ToB:           map[uint64]int{2: 2},
		FromAToB: map[uint64]map[uint64]int{
			1: {2: 2},
		},
	}
	nodeLog := NewFromSnapshot(input)

	input.FromA[1] = 99
	input.ToB[2] = 99
	input.FromAToB[1][2] = 99

	require.Equal(t, want, nodeLog.Snapshot())

	output := nodeLog.Snapshot()
	output.FromA[1] = 88
	output.ToB[2] = 88
	output.FromAToB[1][2] = 88

	require.Equal(t, want, nodeLog.Snapshot())
}
