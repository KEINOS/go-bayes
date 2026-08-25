//go:build windows

package bayes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyncDirectory(t *testing.T) {
	t.Parallel()

	require.NoError(t, syncDirectory("unused"))
}
