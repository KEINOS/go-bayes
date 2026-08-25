package main

import (
	"io"
	"strings"
	"testing"

	"github.com/KEINOS/go-bayes/bayes"
	"github.com/stretchr/testify/require"
)

func TestReadIris_rejectsInvalidData(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"empty dataset":       "",
		"invalid measurement": "not-a-number,3.5,1.4,0.2,Iris-setosa\n",
		"missing field":       "5.1,3.5,1.4,Iris-setosa\n",
		"missing species":     "5.1,3.5,1.4,0.2,\n",
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := readIris(strings.NewReader(input))
			if err == nil {
				t.Fatal("readIris() error = nil, want an error")
			}
		})
	}
}

func TestRun_missingIrisData(t *testing.T) {
	t.Parallel()

	err := runWithDependencies(t.Output(), nil, bayes.MemoryStorage)
	require.ErrorContains(t, err, "load embedded Iris data")
}

func TestRun_predictorCreationFailure(t *testing.T) {
	t.Parallel()

	err := runWithDependencies(t.Output(), rawIrisData, bayes.Storage(99))
	require.ErrorContains(t, err, "create predictor")
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func TestRun_writeFailure(t *testing.T) {
	t.Parallel()

	err := run(failingWriter{})
	require.ErrorContains(t, err, "write training summary")
	require.ErrorIs(t, err, io.ErrClosedPipe)
}
