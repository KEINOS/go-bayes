package main

import (
	"strings"
	"testing"
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
