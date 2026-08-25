//go:build !windows

package bayes

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSyncDirectory(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "unknown_path")

	err := syncDirectory(dir)
	require.Error(t, err,
		"non-existent directory should error")
}
