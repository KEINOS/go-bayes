package class

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSalt_golden(t *testing.T) {
	a, err := GetSalt(int(1))
	require.NoError(t, err)

	b, err := GetSalt(int(2))
	require.NoError(t, err)

	require.Equal(t, a, b, "same type should have same salt")
	assert.IsType(t, byte(100), a, "salt should be a byte or uint8")
	assert.IsType(t, uint8(100), b, "salt should be a byte or uint8")
}

func TestGetSalt_unsupported_type(t *testing.T) {
	a, err := GetSalt(make(chan int))

	require.Error(t, err, "unsupported type should return error")
	require.Zero(t, a, "it should return 0 for unsupported type")

	require.Contains(t, err.Error(), "unsupported type:")
}
