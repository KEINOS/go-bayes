package main

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func TestRun_reportsWriteFailure(t *testing.T) {
	t.Parallel()

	err := run(failingWriter{})

	require.ErrorContains(t, err, "write prediction")
	require.ErrorIs(t, err, io.ErrClosedPipe)
}
