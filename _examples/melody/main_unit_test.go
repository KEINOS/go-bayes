package main

import (
	"io"
	"testing"

	"github.com/KEINOS/go-bayes/bayes"
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

func TestRun_predictorCreationFailure(t *testing.T) {
	t.Parallel()

	err := runWithStorage(t.Output(), bayes.Storage(99))
	require.ErrorContains(t, err, "create predictor")
}
