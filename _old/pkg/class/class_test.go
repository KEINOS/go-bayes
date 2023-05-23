package class

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew_err(t *testing.T) {
	c, err := New(nil)

	require.Error(t, err)
	require.Nil(t, c)
	require.Contains(t, err.Error(), "failed to instantiate Class object")
}

func TestNew_same_value_different_type(t *testing.T) {
	for _, tt := range []struct {
		A any
		B any
	}{
		{int(1), int64(1)},
		{"foo", []byte("foo")},
		{false, 0},
		{float64(1.0), float32(1.0)},
	} {
		a, err := New(tt.A)
		require.NoError(t, err)

		b, err := New(tt.B)
		require.NoError(t, err)

		require.NotEqual(t, a.ID, b.ID,
			"same value but different type should have different ID. A: %T, B: %T", a.Raw, b.Raw)
	}
}
