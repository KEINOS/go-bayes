package main

import (
	"io"
	"strings"
	"testing"

	"github.com/KEINOS/go-bayes/bayes"
	"github.com/stretchr/testify/require"
)

func TestReadMushrooms_rejectsInvalidData(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"empty dataset": "",
		"empty feature": "p,,s,n,t,p,f,c,n,k,e,e,s,s,w,w,p,w,o,p,k,s,u\n",
		"long feature":  "p,xx,s,n,t,p,f,c,n,k,e,e,s,s,w,w,p,w,o,p,k,s,u\n",
		"missing field": "p,x,s,n\n",
		"unknown class": "q,x,s,n,t,p,f,c,n,k,e,e,s,s,w,w,p,w,o,p,k,s,u\n",
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := readMushrooms(strings.NewReader(input))
			if err == nil {
				t.Fatal("readMushrooms() error = nil, want an error")
			}
		})
	}
}

func TestRun_missingMushroomData(t *testing.T) {
	t.Parallel()

	err := runWithDependencies(t.Output(), nil, bayes.MemoryStorage)
	require.ErrorContains(t, err, "load embedded Mushroom data")
}

func TestRun_predictorCreationFailure(t *testing.T) {
	t.Parallel()

	err := runWithDependencies(t.Output(), rawMushroomData, bayes.Storage(99))
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
